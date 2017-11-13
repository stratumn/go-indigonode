// Copyright © 2017  Stratumn SAS
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"github.com/chzyer/readline"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jbenet/go-base58"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/stratumn/alice/grpc/ext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"

	maddr "gx/ipfs/QmXY77cVe7rVRQXZZQRioukUM7aRW3BTcAgJe12MCtb3Ji/go-multiaddr"
)

/*
Assuming server reflection is enabled on the gRPC server, the functions in this
file attempt to create commands dynamically by inspecting the available
services. It also looks for Alice protobuf extensions to get help strings for
the commands.

At the moment it only supports string and bool fields in the request message,
and only works with unary requests and server streams.

This is a bit complex, have a look at the photoreflect package documentation.
*/

// reflectAPI uses reflection to automatically create commands from a gRPC
// server.
func reflectAPI(ctx context.Context, conn *grpc.ClientConn, cons *Console) ([]Cmd, error) {
	cons.Debugln("Reflecting API commands...")

	stub := grpc_reflection_v1alpha.NewServerReflectionClient(conn)
	client := grpcreflect.NewClient(ctx, stub)

	services, err := client.ListServices()
	if err != nil {
		return nil, err
	}

	var cmds []Cmd

	for _, name := range services {
		// Ignore the server reflection service.
		if strings.HasPrefix(name, "grpc.reflection") {
			cons.Debugf("Ignoring %q.\n", name)
			continue
		}

		c, err := reflectService(client, conn, name, cons)
		if err != nil {
			return nil, err
		}

		cmds = append(cmds, c...)
	}

	client.Reset()

	cons.Debugln("Reflected API commands.")

	return cmds, nil
}

// reflectService automatically creates commands for a service of a gRPC
// server.
func reflectService(client *grpcreflect.Client, conn *grpc.ClientConn, name string, cons *Console) ([]Cmd, error) {
	cons.Debugf("Reflecting service %q...\n", name)

	serv, err := client.ResolveService(name)
	if err != nil {
		return nil, err
	}

	var cmds []Cmd
	methods := serv.GetMethods()

	for _, method := range methods {
		c, err := reflectMethod(conn, method)
		if err != nil {
			cons.Warningf("Could not reflect %q: %s.\n", method.GetFullyQualifiedName(), err)
			continue
		}
		if c != nil {
			cmds = append(cmds, c)
		}
	}

	cons.Debugf("Reflected service %q.\n", name)
	return cmds, nil
}

// reflectMethod automatically creates a command for a method of a gRPC server.
func reflectMethod(conn *grpc.ClientConn, method *desc.MethodDescriptor) (Cmd, error) {
	if method.IsClientStreaming() {
		return nil, errors.WithStack(ErrUnsupportedReflectType)
	}

	serviceName := strings.ToLower(method.GetService().GetName())
	methodName := strings.ToLower(method.GetName())
	name := methodName

	if serviceName != methodName {
		name = serviceName + "-" + name
	}

	c := BasicCmd{
		Name:  name,
		Use:   name,
		Short: reflectShort(method),
	}

	fields := method.GetInputType().GetFields()
	required, optional, err := filterDynamicFields(fields)
	if err != nil {
		return nil, err
	}

	for _, field := range required {
		c.Use += " <" + reflectFieldDesc(field) + ">"
	}

	c.Flags = func() *pflag.FlagSet {
		return reflectFlags(method.GetFullyQualifiedName(), optional)
	}

	c.Exec = func(ctx context.Context, cli *CLI, args []string, flags *pflag.FlagSet) error {
		return reflectExec(ctx, cli, args, flags, method, required, optional, conn)
	}

	return BasicCmdWrapper{c}, nil
}

// filterDynamicFields filters the required and optional fields of a method.
func filterDynamicFields(fields []*desc.FieldDescriptor) ([]*desc.FieldDescriptor, []*desc.FieldDescriptor, error) {
	var required []*desc.FieldDescriptor
	var optional []*desc.FieldDescriptor

	for _, field := range fields {
		if field.IsRepeated() {
			return nil, nil, errors.WithStack(ErrUnsupportedReflectType)
		}

		switch field.GetType() {
		case descriptor.FieldDescriptorProto_TYPE_STRING:
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
		case descriptor.FieldDescriptorProto_TYPE_BYTES:
		case descriptor.FieldDescriptorProto_TYPE_INT64:
		case descriptor.FieldDescriptorProto_TYPE_UINT32:
		default:
			return nil, nil, errors.WithStack(ErrUnsupportedReflectType)
		}

		if reflectFieldRequired(field) {
			required = append(required, field)
			continue
		}

		optional = append(optional, field)
	}

	return required, optional, nil
}

// reflectShort uses reflection to get the short string of a command.
func reflectShort(method *desc.MethodDescriptor) string {
	options := method.GetMethodOptions()

	if options != nil {
		ex, err := proto.GetExtension(options, ext.E_MethodDesc)
		if err == nil {
			return *ex.(*string)
		}
	}

	return fmt.Sprintf(
		"Call the %s method of the %s service",
		method.GetName(),
		method.GetService().GetFullyQualifiedName(),
	)
}

// reflectFlags uses reflection to create a flag set.
func reflectFlags(name string, fields []*desc.FieldDescriptor) *pflag.FlagSet {
	flags := pflag.NewFlagSet(name, pflag.ContinueOnError)

	flags.Bool("no-truncate", false, "Disable truncating rows")
	flags.Bool("no-borders", false, "Disable table borders")

	for _, field := range fields {
		fieldName := field.GetName()
		use := reflectFieldDesc(field)

		switch field.GetType() {
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			flags.String(fieldName, "", use)
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			flags.Bool(fieldName, false, use)
		case descriptor.FieldDescriptorProto_TYPE_BYTES:
			flags.String(fieldName, "", use)
		case descriptor.FieldDescriptorProto_TYPE_INT64:
			flags.Int64(fieldName, 0, use)
		case descriptor.FieldDescriptorProto_TYPE_UINT32:
			flags.Uint32(fieldName, 0, use)
		}
	}

	return flags
}

// reflectFieldRequired uses reflection to check if a field is required.
func reflectFieldRequired(field *desc.FieldDescriptor) bool {
	options := field.GetFieldOptions()
	if options == nil {
		return false
	}

	ex, err := proto.GetExtension(options, ext.E_FieldRequired)
	if err != nil {
		return false
	}

	return *ex.(*bool)
}

// reflectFieldDesc uses reflection to get a field's description.
func reflectFieldDesc(field *desc.FieldDescriptor) string {
	options := field.GetFieldOptions()
	if options != nil {
		ex, err := proto.GetExtension(options, ext.E_FieldDesc)
		if err == nil {
			return *ex.(*string)
		}
	}

	return field.GetName() + " field of the API request"
}

// reflectExec uses reflection to execute a command.
func reflectExec(
	ctx context.Context,
	cli *CLI,
	args []string,
	flags *pflag.FlagSet,
	method *desc.MethodDescriptor,
	required []*desc.FieldDescriptor,
	optional []*desc.FieldDescriptor,
	conn *grpc.ClientConn,
) error {
	argsLen := len(args)
	reqLen := len(required)

	if argsLen < reqLen {
		return NewUseError("missing argument(s)")
	}
	if argsLen > reqLen {
		return NewUseError("unexpected argument(s): " + strings.Join(args[reqLen:], " "))
	}

	req := dynamic.NewMessage(method.GetInputType())

	if err := setDynamicFlags(req, optional, flags); err != nil {
		return err
	}
	if err := setDynamicArgs(req, required, args); err != nil {
		return err
	}

	stub := grpcdynamic.NewStub(conn)
	w := cli.Console()

	to, err := time.ParseDuration(cli.Config().DialTimeout)
	if err != nil {
		return errors.WithStack(err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, to)
	defer cancel()

	if method.IsServerStreaming() {
		ss, err := stub.InvokeRpcServerStream(reqCtx, method, req)
		if err != nil {
			return errors.WithStack(err)
		}

		noTrunc, err := flags.GetBool("no-truncate")
		if err != nil {
			return errors.WithStack(err)
		}

		noBord, err := flags.GetBool("no-borders")
		if err != nil {
			return errors.WithStack(err)
		}

		return printDynamicStream(w, ss, !noTrunc, !noBord)
	}

	res, err := stub.InvokeRpc(reqCtx, method, req)
	if err != nil {
		return errors.WithStack(err)
	}

	return printDynamicMsg(w, res.(*dynamic.Message))
}

// setDynamicFlags sets the dynamic request values from flags.
func setDynamicFlags(req *dynamic.Message, fields []*desc.FieldDescriptor, flags *pflag.FlagSet) error {
	for _, field := range fields {
		fieldName := field.GetName()

		switch field.GetType() {
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			value, err := flags.GetString(fieldName)
			if err != nil {
				return errors.WithStack(err)
			}
			req.SetFieldByName(fieldName, value)
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			value, err := flags.GetBool(fieldName)
			if err != nil {
				return errors.WithStack(err)
			}
			req.SetFieldByName(fieldName, value)
		case descriptor.FieldDescriptorProto_TYPE_BYTES:
			value, err := flags.GetString(fieldName)
			if err != nil {
				return errors.WithStack(err)
			}
			if value == "" {
				continue
			}

			b, err := reflectBytesValue(field, value)
			if err != nil {
				return err
			}

			req.SetFieldByName(fieldName, b)
		case descriptor.FieldDescriptorProto_TYPE_INT64:
			// TODO: handle duration.
			value, err := flags.GetInt64(fieldName)
			if err != nil {
				return errors.WithStack(err)
			}
			req.SetFieldByName(fieldName, value)
		case descriptor.FieldDescriptorProto_TYPE_UINT32:
			value, err := flags.GetUint32(fieldName)
			if err != nil {
				return errors.WithStack(err)
			}
			req.SetFieldByName(fieldName, value)
		}
	}

	return nil
}

// setDymanicArgs sets the dynamic request values from arguments.
func setDynamicArgs(req *dynamic.Message, fields []*desc.FieldDescriptor, args []string) error {
	for i, field := range fields {
		fieldName := field.GetName()
		value := args[i]

		switch field.GetType() {
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			req.SetFieldByName(fieldName, value)
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			switch value {
			case "true":
				req.SetFieldByName(fieldName, true)
			case "false":
				req.SetFieldByName(fieldName, false)
			default:
				return NewUseError("invalid bool value: " + value)
			}
		case descriptor.FieldDescriptorProto_TYPE_BYTES:
			b, err := reflectBytesValue(field, value)
			if err != nil {
				return err
			}
			req.SetFieldByName(fieldName, b)
		case descriptor.FieldDescriptorProto_TYPE_INT64:
			// TODO: handle duration.
			i, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return errors.WithStack(err)
			}
			req.SetFieldByName(fieldName, i)
		case descriptor.FieldDescriptorProto_TYPE_UINT32:
			i, err := strconv.ParseUint(value, 10, 32)
			if err != nil {
				return errors.WithStack(err)
			}
			req.SetFieldByName(fieldName, i)
		}
	}

	return nil
}

// printDynamicMsg prints a unary dynamic message received from the API.
func printDynamicMsg(w io.Writer, msg *dynamic.Message) error {
	tw := new(tabwriter.Writer)
	tw.Init(w, 0, 8, 2, ' ', 0)

	for _, field := range msg.GetKnownFields() {
		name := field.GetName()
		label := strings.ToUpper(name)
		value := prettyDynamicVal(msg, field)
		fmt.Fprintf(tw, "%s\t%v\n", label, value)
	}

	return errors.WithStack(tw.Flush())
}

// printDynamicStream prints a dynamic stream received from the API.
func printDynamicStream(w io.Writer, ss *grpcdynamic.ServerStream, truncate, borders bool) error {
	var b bytes.Buffer
	var numCols int

	tw := new(tabwriter.Writer)
	tw.Init(&b, 0, 8, 1, ' ', 0)

	// Used to check if it should print the table header.
	firstRow := true

	for {
		res, err := ss.RecvMsg()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.WithStack(err)
		}

		msg := res.(*dynamic.Message)
		fields := msg.GetKnownFields()
		numCols = len(fields)

		if firstRow && numCols > 1 {
			for i, field := range fields {
				label := strings.ToUpper(field.GetName())

				if i < numCols-1 {
					if borders {
						fmt.Fprintf(tw, "%s\t| ", label)
					} else {
						fmt.Fprintf(tw, "%s\t", label)
					}
				} else {
					fmt.Fprintf(tw, "%s", label)
				}
			}

			fmt.Fprintln(tw, "")

			if borders {
				for range fields[:numCols-1] {
					fmt.Fprint(tw, "\t+")
				}

				fmt.Fprintln(tw, "\t")
			}
		}

		for i, field := range fields {
			value := prettyDynamicVal(msg, field)

			if borders && i < numCols-1 {
				fmt.Fprintf(tw, "%s\t| ", value)
			} else {
				fmt.Fprintf(tw, "%s\t", value)
			}
		}

		fmt.Fprintln(tw, "")
		firstRow = false

	}

	if err := tw.Flush(); err != nil {
		return errors.WithStack(err)
	}

	// Truncate output.
	width := readline.GetScreenWidth()

	for i, row := range strings.Split(b.String(), "\n") {
		if borders && numCols > 1 && i == 1 {
			row = strings.Replace(row, " ", "-", width)
			l := len(row) - 1
			if l > width {
				l = width
			}
			row = row[:l]
			fmt.Fprintln(w, row)
			continue
		}

		row := strings.TrimSpace(row)
		runes := []rune(row)

		if truncate && width > 0 && len(runes) > width {
			fmt.Fprintln(w, string(runes[:width-3])+"...")
		} else if row != "" {
			fmt.Fprintln(w, row)
		}
	}

	return nil
}

// prettyDynamicVal returns a strings representation of a value returns by the
// API.
func prettyDynamicVal(msg *dynamic.Message, field *desc.FieldDescriptor) string {
	value := msg.GetFieldByName(field.GetName())

	if enum := field.GetEnumType(); enum != nil {
		return enum.FindValueByNumber(value.(int32)).GetName()
	}

	switch v := value.(type) {
	case []byte:
		options := field.GetFieldOptions()
		if options != nil {
			// Base58.
			ex, err := proto.GetExtension(options, ext.E_FieldBase58)
			if err == nil && *ex.(*bool) {
				return base58.Encode(v)
			}

			// Multiaddr.
			ex, err = proto.GetExtension(options, ext.E_FieldMultiaddr)
			if err == nil && *ex.(*bool) {
				addr, err := maddr.NewMultiaddrBytes(v)
				if err == nil {
					return addr.String()
				}
			}
		}

		return hex.EncodeToString(v)

	case int64:
		options := field.GetFieldOptions()
		if options != nil {
			// Duration.
			ex, err := proto.GetExtension(options, ext.E_FieldDuration)
			if err == nil && *ex.(*bool) {
				return fmt.Sprintf("%v", time.Duration(v))
			}
		}
		return fmt.Sprintf("%d", v)

	case uint64:
		options := field.GetFieldOptions()
		if options != nil {
			// Bytesize.
			ex, err := proto.GetExtension(options, ext.E_FieldBytesize)
			if err == nil && *ex.(*bool) {
				return bytefmt.ByteSize(v)
			}

			// Byterate.
			ex, err = proto.GetExtension(options, ext.E_FieldByterate)
			if err == nil && *ex.(*bool) {
				return fmt.Sprintf("%s/s", bytefmt.ByteSize(v))
			}
		}
		return fmt.Sprintf("%d", v)

	default:
		return fmt.Sprintf("%v", v)
	}
}

// reflectBytesValue sets the value of a field of type bytes from a string.
func reflectBytesValue(field *desc.FieldDescriptor, value string) ([]byte, error) {
	var b []byte
	var err error

	options := field.GetFieldOptions()

	if options != nil {
		// Base58.
		ex, err := proto.GetExtension(options, ext.E_FieldBase58)
		if err == nil && *ex.(*bool) {
			b = base58.Decode(value)
			if len(b) < 1 {
				return nil, errors.WithStack(ErrInvalidBase58)
			}
		}

		// Multiaddr.
		if len(b) < 1 {
			ex, err = proto.GetExtension(options, ext.E_FieldMultiaddr)
			if err == nil && *ex.(*bool) {
				addr, err := maddr.NewMultiaddr(value)
				if err != nil {
					return nil, errors.WithStack(err)
				}

				b = addr.Bytes()
			}
		}
	}

	// Hex.
	if len(b) < 1 {
		b, err = hex.DecodeString(value)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return b, nil
}
