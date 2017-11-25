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
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"github.com/chzyer/readline"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	base58 "github.com/jbenet/go-base58"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/stratumn/alice/grpc/ext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"

	ma "gx/ipfs/QmXY77cVe7rVRQXZZQRioukUM7aRW3BTcAgJe12MCtb3Ji/go-multiaddr"
)

// ReflectFieldDesc returns a description a the gRPC field.
//
// It will look for the field description extension if available.
func ReflectFieldDesc(d *desc.FieldDescriptor) string {
	opts := d.GetFieldOptions()
	if opts != nil {
		ex, err := proto.GetExtension(opts, ext.E_FieldDesc)
		if err == nil {
			return *ex.(*string)
		}
	}

	return strings.Replace(d.GetName(), "_", " ", -1)
}

// ReflectMethodDesc returns a description a the gRPC method.
//
// It will look for the method description extension if available.
func ReflectMethodDesc(d *desc.MethodDescriptor) string {
	opts := d.GetMethodOptions()

	if opts != nil {
		ex, err := proto.GetExtension(opts, ext.E_MethodDesc)
		if err == nil {
			return *ex.(*string)
		}
	}

	return fmt.Sprintf(
		"Call the %s method of the %s service",
		d.GetName(),
		d.GetService().GetFullyQualifiedName(),
	)
}

// ReflectFieldRequired returns whether a gRPC field is required.
//
// It will look for the field required extension if available.
func ReflectFieldRequired(d *desc.FieldDescriptor) bool {
	opts := d.GetFieldOptions()
	if opts == nil {
		return false
	}

	ex, err := proto.GetExtension(opts, ext.E_FieldRequired)
	if err != nil {
		return false
	}

	return *ex.(*bool)
}

// Reflector reflects gRPC fields.
type Reflector interface {
	// Supports returns whether it can handle this field.
	Supports(*desc.FieldDescriptor) bool
}

// ArgReflector reflects values for a gRPC request from a command argument.
type ArgReflector interface {
	Reflector

	// Parse parses the value for the field from a string.
	Parse(*desc.FieldDescriptor, string) (interface{}, error)
}

// FlagReflector reflects values for a gRPC request from a command flag.
type FlagReflector interface {
	Reflector

	// Flag adds a flag for the value to a flag set.
	Flag(*desc.FieldDescriptor, *pflag.FlagSet)

	// ParseFlag parses the value of the flag.
	ParseFlag(*desc.FieldDescriptor, *pflag.FlagSet) (interface{}, error)
}

// ResponseReflector reflects values of a gRPC request.
type ResponseReflector interface {
	Reflector

	// Pretty returns a human friendly representation of the value.
	Pretty(*desc.FieldDescriptor, interface{}) (string, error)
}

// ReflectChecker checks if a field is supported by a reflector
type ReflectChecker func(*desc.FieldDescriptor) bool

// ReflectEncoder encodes a value to a string
type ReflectEncoder func(*desc.FieldDescriptor, interface{}) (string, error)

// ReflectDecoder decodes a value from a string.
type ReflectDecoder func(*desc.FieldDescriptor, string) (interface{}, error)

// GenericReflector is a generic reflector that covers most use cases.
type GenericReflector struct {
	// Zero is the zero value of the primitive type in a protocol buffer.
	Zero interface{}

	Checker ReflectChecker
	Encoder ReflectEncoder
	Decoder ReflectDecoder
}

// Supports returns whether it can handle this type of field.
func (r GenericReflector) Supports(d *desc.FieldDescriptor) bool {
	return r.Checker(d)
}

// Parse parses the value for the field from a string.
func (r GenericReflector) Parse(d *desc.FieldDescriptor, s string) (interface{}, error) {
	s = strings.TrimSpace(s)

	if d.IsRepeated() {
		vals := strings.Split(s, ",")
		l := len(vals)
		t := reflect.SliceOf(reflect.TypeOf(r.Zero))
		res := reflect.MakeSlice(t, l, l)

		for i, v := range vals {
			decoded, err := r.Decoder(d, strings.TrimSpace(v))
			if err != nil {
				return reflect.MakeSlice(t, 0, 0).Interface(), err
			}
			res.Index(i).Set(reflect.ValueOf(decoded))
		}

		return res.Interface(), nil
	}

	return r.Decoder(d, s)
}

// Flag adds a flag for the value to a flag set.
func (r GenericReflector) Flag(d *desc.FieldDescriptor, f *pflag.FlagSet) {
	help := ReflectFieldDesc(d)

	if d.IsRepeated() {
		f.StringSlice(d.GetName(), []string{}, help)
		return
	}

	f.String(d.GetName(), "", help)
}

// ParseFlag parses the value of the flag.
func (r GenericReflector) ParseFlag(d *desc.FieldDescriptor, f *pflag.FlagSet) (interface{}, error) {
	if d.IsRepeated() {
		v, err := f.GetStringSlice(d.GetName())
		if err != nil {
			return nil, errors.WithStack(err)
		}

		l := len(v)
		t := reflect.SliceOf(reflect.TypeOf(r.Zero))
		res := reflect.MakeSlice(t, l, l)

		for i, s := range v {
			decoded, err := r.Decoder(d, strings.TrimSpace(s))
			if err != nil {
				return reflect.MakeSlice(t, 0, 0).Interface(), err
			}
			res.Index(i).Set(reflect.ValueOf(decoded))
		}

		return res.Interface(), nil
	}

	v, err := f.GetString(d.GetName())
	if err != nil {
		return "", errors.WithStack(err)
	}
	if v == "" {
		return r.Zero, nil
	}
	return r.Decoder(d, strings.TrimSpace(v))
}

// Pretty returns a human friendly representation of the value.
func (r GenericReflector) Pretty(d *desc.FieldDescriptor, v interface{}) (string, error) {
	if d.IsRepeated() {
		v := reflect.ValueOf(v)
		l := v.Len()
		s := make([]string, l)

		for i := range s {
			encoded, err := r.Encoder(d, v.Index(i).Interface())
			if err != nil {
				return "", err
			}

			s[i] = encoded
		}

		return strings.Join(s, ","), nil
	}

	return r.Encoder(d, v)
}

// NewStringReflector creates a new string reflector.
func NewStringReflector() Reflector {
	return GenericReflector{
		Zero: "",
		Checker: func(d *desc.FieldDescriptor) bool {
			return d.GetType() == descriptor.FieldDescriptorProto_TYPE_STRING
		},
		Encoder: func(d *desc.FieldDescriptor, v interface{}) (string, error) {
			return v.(string), nil
		},
		Decoder: func(d *desc.FieldDescriptor, s string) (interface{}, error) {
			return s, nil
		},
	}
}

// NewBoolReflector creates a new bool reflector.
func NewBoolReflector() Reflector {
	return GenericReflector{
		Zero: false,
		Checker: func(d *desc.FieldDescriptor) bool {
			return d.GetType() == descriptor.FieldDescriptorProto_TYPE_BOOL
		},
		Encoder: func(d *desc.FieldDescriptor, v interface{}) (string, error) {
			return fmt.Sprintf("%v", v), nil
		},
		Decoder: func(d *desc.FieldDescriptor, s string) (interface{}, error) {
			switch strings.ToLower(s) {
			case "true":
				return true, nil
			case "false":
				return false, nil
			}

			return false, errors.WithStack(ErrParse)
		},
	}
}

// NewUint32Reflector creates a new uint32 reflector.
func NewUint32Reflector() Reflector {
	return GenericReflector{
		Zero: uint32(0),
		Checker: func(d *desc.FieldDescriptor) bool {
			return d.GetType() == descriptor.FieldDescriptorProto_TYPE_UINT32
		},
		Encoder: func(d *desc.FieldDescriptor, v interface{}) (string, error) {
			return fmt.Sprintf("%d", v), nil
		},
		Decoder: func(d *desc.FieldDescriptor, s string) (interface{}, error) {
			i, err := strconv.ParseUint(s, 10, 32)
			return uint32(i), errors.WithStack(err)
		},
	}
}

// NewBytesReflector creates a new bytes reflector.
func NewBytesReflector() Reflector {
	return GenericReflector{
		Zero: []byte{},
		Checker: func(d *desc.FieldDescriptor) bool {
			return d.GetType() == descriptor.FieldDescriptorProto_TYPE_BYTES
		},
		Encoder: func(d *desc.FieldDescriptor, v interface{}) (string, error) {
			return hex.EncodeToString(v.([]byte)), nil
		},
		Decoder: func(d *desc.FieldDescriptor, s string) (interface{}, error) {
			v, err := hex.DecodeString(s)
			if err != nil {
				return []byte{}, errors.WithStack(err)
			}

			return v, nil
		},
	}
}

// NewEnumReflector creates a new enum reflector.
func NewEnumReflector() Reflector {
	return GenericReflector{
		Zero: int32(0),
		Checker: func(d *desc.FieldDescriptor) bool {
			return d.GetType() == descriptor.FieldDescriptorProto_TYPE_ENUM
		},
		Encoder: func(d *desc.FieldDescriptor, v interface{}) (string, error) {
			t := d.GetEnumType()
			return t.FindValueByNumber(v.(int32)).GetName(), nil
		},
		Decoder: func(d *desc.FieldDescriptor, s string) (interface{}, error) {
			s = strings.ToLower(s)
			t := d.GetEnumType()
			for _, vt := range t.GetValues() {
				if s == strings.ToLower(vt.GetName()) {
					return vt.GetNumber(), nil
				}
			}

			return int32(0), errors.WithStack(ErrParse)
		},
	}
}

// NewDurationReflector creates a new duration reflector.
func NewDurationReflector() Reflector {
	return GenericReflector{
		Zero: int64(0),
		Checker: func(d *desc.FieldDescriptor) bool {
			if d.GetType() != descriptor.FieldDescriptorProto_TYPE_INT64 {
				return false
			}

			opts := d.GetFieldOptions()
			if opts == nil {
				return false
			}

			ex, err := proto.GetExtension(opts, ext.E_FieldDuration)

			return err == nil && *ex.(*bool)
		},
		Encoder: func(d *desc.FieldDescriptor, v interface{}) (string, error) {
			return time.Duration(v.(int64)).String(), nil
		},
		Decoder: func(d *desc.FieldDescriptor, s string) (interface{}, error) {
			v, err := time.ParseDuration(s)
			return int64(v), errors.WithStack(err)
		},
	}
}

// NewBase58Reflector creates a new base58 reflector.
func NewBase58Reflector() Reflector {
	return GenericReflector{
		Zero: []byte{},
		Checker: func(d *desc.FieldDescriptor) bool {
			if d.GetType() != descriptor.FieldDescriptorProto_TYPE_BYTES {
				return false
			}

			opts := d.GetFieldOptions()
			if opts == nil {
				return false
			}

			ex, err := proto.GetExtension(opts, ext.E_FieldBase58)

			return err == nil && *ex.(*bool)
		},
		Encoder: func(d *desc.FieldDescriptor, v interface{}) (string, error) {
			return base58.Encode(v.([]byte)), nil
		},
		Decoder: func(d *desc.FieldDescriptor, s string) (interface{}, error) {
			v := base58.Decode(s)
			if len(v) < 1 {
				return []byte{}, errors.WithStack(ErrParse)
			}

			return v, nil
		},
	}
}

// NewBytesizeReflector creates a new bytesize reflector.
func NewBytesizeReflector() Reflector {
	return GenericReflector{
		Zero: uint64(0),
		Checker: func(d *desc.FieldDescriptor) bool {
			if d.GetType() != descriptor.FieldDescriptorProto_TYPE_UINT64 {
				return false
			}

			opts := d.GetFieldOptions()
			if opts == nil {
				return false
			}

			ex, err := proto.GetExtension(opts, ext.E_FieldBytesize)

			return err == nil && *ex.(*bool)
		},
		Encoder: func(d *desc.FieldDescriptor, v interface{}) (string, error) {
			return bytefmt.ByteSize(v.(uint64)), nil
		},
		Decoder: func(d *desc.FieldDescriptor, s string) (interface{}, error) {
			v, err := bytefmt.ToBytes(s)
			return v, errors.WithStack(err)
		},
	}
}

// NewByterateReflector creates a new byterate reflector.
func NewByterateReflector() Reflector {
	return GenericReflector{
		Zero: uint64(0),
		Checker: func(d *desc.FieldDescriptor) bool {
			if d.GetType() != descriptor.FieldDescriptorProto_TYPE_UINT64 {
				return false
			}

			opts := d.GetFieldOptions()
			if opts == nil {
				return false
			}

			ex, err := proto.GetExtension(opts, ext.E_FieldByterate)

			return err == nil && *ex.(*bool)
		},
		Encoder: func(d *desc.FieldDescriptor, v interface{}) (string, error) {
			return bytefmt.ByteSize(v.(uint64)) + "/s", nil
		},
		Decoder: func(d *desc.FieldDescriptor, s string) (interface{}, error) {
			s = strings.TrimSuffix(s, "/s")
			v, err := bytefmt.ToBytes(s)
			return v, errors.WithStack(err)
		},
	}
}

// NewMaddrReflector creates a new multiaddr reflector.
func NewMaddrReflector() Reflector {
	return GenericReflector{
		Zero: []byte{},
		Checker: func(d *desc.FieldDescriptor) bool {
			if d.GetType() != descriptor.FieldDescriptorProto_TYPE_BYTES {
				return false
			}

			opts := d.GetFieldOptions()
			if opts == nil {
				return false
			}

			ex, err := proto.GetExtension(opts, ext.E_FieldMultiaddr)

			return err == nil && *ex.(*bool)
		},
		Encoder: func(d *desc.FieldDescriptor, v interface{}) (string, error) {
			maddr, err := ma.NewMultiaddrBytes(v.([]byte))
			if err != nil {
				return "", errors.WithStack(err)
			}

			return maddr.String(), nil
		},
		Decoder: func(d *desc.FieldDescriptor, s string) (interface{}, error) {
			v, err := ma.NewMultiaddr(s)
			if err != nil {
				return []byte{}, errors.WithStack(err)
			}

			return v.Bytes(), nil
		},
	}
}

// DefReflectors are the default reflectors used by NewServerReflector.
var DefReflectors = []Reflector{
	NewDurationReflector(),
	NewBase58Reflector(),
	NewBytesizeReflector(),
	NewByterateReflector(),
	NewMaddrReflector(),
	NewStringReflector(),
	NewBoolReflector(),
	NewUint32Reflector(),
	NewBytesReflector(),
	NewEnumReflector(),
}

// ServerReflector reflects commands from a gRPC server.
//
// The server must have the reflection service enabled.
type ServerReflector struct {
	cons           *Console
	argReflectors  []ArgReflector
	flagReflectors []FlagReflector
	resReflectors  []ResponseReflector
	width          int
}

// NewServerReflector reflects commands from a gRPC server.
//
// When reflecting fields, the first reflector that supports it is used. This
// means that the more specific reflectors should be passed first.
//
// If no reflectors are given, DefReflectors is used.
func NewServerReflector(cons *Console, termWidth int, reflectors ...Reflector) ServerReflector {
	if len(reflectors) < 1 {
		reflectors = DefReflectors
	}

	if termWidth == 0 {
		termWidth = readline.GetScreenWidth()
	}

	r := ServerReflector{cons: cons, width: termWidth}

	for _, reflector := range reflectors {
		if v, ok := reflector.(ArgReflector); ok {
			r.argReflectors = append(r.argReflectors, v)
		}

		if v, ok := reflector.(FlagReflector); ok {
			r.flagReflectors = append(r.flagReflectors, v)
		}

		if v, ok := reflector.(ResponseReflector); ok {
			r.resReflectors = append(r.resReflectors, v)
		}
	}

	return r
}

// Reflect reflects the command of a server and returns commands for them.
func (r ServerReflector) Reflect(ctx context.Context, conn *grpc.ClientConn) ([]Cmd, error) {
	r.cons.Debugln("Reflecting API commands...")

	stub := grpc_reflection_v1alpha.NewServerReflectionClient(conn)
	c := grpcreflect.NewClient(ctx, stub)
	defer c.Reset()

	servNames, err := c.ListServices()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var cmds []Cmd
	for _, d := range r.getServiceDescs(c, servNames) {
		cmds = append(cmds, r.reflectService(conn, d)...)
	}

	r.cons.Debugln("Reflected API commands.")

	return cmds, err
}

// getServicesDescs gets the service descriptors for the given service names.
func (r ServerReflector) getServiceDescs(c *grpcreflect.Client, servNames []string) []*desc.ServiceDescriptor {
	var descs []*desc.ServiceDescriptor

	for _, name := range servNames {
		// Ignore the server reflection service.
		if strings.HasPrefix(name, "grpc.reflection") {
			r.cons.Debugf("Ignoring %q.\n", name)
			continue
		}

		d, err := c.ResolveService(name)
		if err != nil {
			if err != nil {
				r.cons.Warningf("Could not get service descriptor for %q: %s\n", name, err)
				continue
			}
			continue
		}

		descs = append(descs, d)
	}

	return descs
}

// reflectService reflect commands for the given service descriptor.
func (r ServerReflector) reflectService(conn *grpc.ClientConn, d *desc.ServiceDescriptor) []Cmd {
	methodDescs := d.GetMethods()

	var cmds []Cmd
	for _, methodDesc := range methodDescs {
		c, err := r.reflectMethod(conn, methodDesc)
		if err != nil {
			r.cons.Warningf("Could not reflect %q: %s.\n", methodDesc.GetFullyQualifiedName(), err)
			continue
		}

		if c != nil {
			cmds = append(cmds, c)
		}
	}

	return cmds
}

// reflectMethods reflect the command for the given methods descriptor.
func (r ServerReflector) reflectMethod(conn *grpc.ClientConn, d *desc.MethodDescriptor) (Cmd, error) {
	if d.IsClientStreaming() {
		return nil, errors.WithStack(ErrUnsupportedReflectType)
	}

	servName := strings.ToLower(d.GetService().GetName())
	methodName := strings.ToLower(d.GetName())
	name := methodName

	if servName != methodName {
		name = servName + "-" + name
	}

	cmd := BasicCmd{
		Name:  name,
		Use:   name,
		Short: ReflectMethodDesc(d),
	}

	for _, f := range d.GetOutputType().GetFields() {
		if r.findResReflector(f) == nil {
			return nil, ErrUnsupportedReflectType
		}
	}

	inputDescs := d.GetInputType().GetFields()

	required, err := r.findRequiredFields(inputDescs)
	if err != nil {
		return nil, err
	}

	for _, f := range required {
		cmd.Use += " <" + ReflectFieldDesc(f) + ">"
	}

	optional, err := r.findOptionalFields(inputDescs)
	if err != nil {
		return nil, err
	}

	cmd.Flags = func() *pflag.FlagSet {
		return r.flags(d.GetFullyQualifiedName(), optional)
	}

	cmd.Exec = func(ctx context.Context, cli CLI, args []string, flags *pflag.FlagSet) error {
		return r.reflectExec(ctx, cli, args, flags, d, required, optional, conn)
	}

	return BasicCmdWrapper{cmd}, nil
}

// findRequiredFields finds all the required fields.
func (r ServerReflector) findRequiredFields(
	descs []*desc.FieldDescriptor,
) ([]*desc.FieldDescriptor, error) {
	var required []*desc.FieldDescriptor

	for _, d := range descs {
		if ReflectFieldRequired(d) {
			if r.findArgReflector(d) == nil {
				return nil, ErrUnsupportedReflectType
			}

			required = append(required, d)
		}
	}

	return required, nil
}

// findOptionalFields finds all the optional fields.
func (r ServerReflector) findOptionalFields(
	descs []*desc.FieldDescriptor,
) ([]*desc.FieldDescriptor, error) {
	var optional []*desc.FieldDescriptor

	for _, d := range descs {
		if !ReflectFieldRequired(d) {
			if r.findFlagReflector(d) == nil {
				return nil, ErrUnsupportedReflectType
			}

			optional = append(optional, d)
		}
	}

	return optional, nil
}

// findArgReflector finds the first argument reflector that supports a field.
func (r ServerReflector) findArgReflector(d *desc.FieldDescriptor) ArgReflector {
	for _, reflector := range r.argReflectors {
		if reflector.Supports(d) {
			return reflector
		}
	}

	return nil
}

// findFlagReflector finds the first flag reflector that supports a field.
func (r ServerReflector) findFlagReflector(d *desc.FieldDescriptor) FlagReflector {
	for _, reflector := range r.flagReflectors {
		if reflector.Supports(d) {
			return reflector
		}
	}

	return nil
}

// findResReflector finds the first response reflector that supports a field.
func (r ServerReflector) findResReflector(d *desc.FieldDescriptor) ResponseReflector {
	for _, reflector := range r.resReflectors {
		if reflector.Supports(d) {
			return reflector
		}
	}

	return nil
}

// flags adds flags to a command.
func (r ServerReflector) flags(name string, descs []*desc.FieldDescriptor) *pflag.FlagSet {
	flags := pflag.NewFlagSet(name, pflag.ContinueOnError)

	flags.Bool("no-truncate", false, "Disable truncating rows")
	flags.Bool("no-borders", false, "Disable table borders")

	for _, d := range descs {
		reflector := r.findFlagReflector(d)
		if reflector == nil {
			r.cons.Errorf(
				"Could not find reflector for %q (this shouldn't happen).\n",
				d.GetName(),
			)
			continue
		}

		reflector.Flag(d, flags)
	}

	return flags
}

// reflectExec executes a command.
func (r ServerReflector) reflectExec(
	ctx context.Context,
	cli CLI,
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

	if err := r.setArgs(req, required, args); err != nil {
		return err
	}
	if err := r.setFlags(req, optional, flags); err != nil {
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

		return r.printStream(w, ss, !noTrunc, !noBord)
	}

	res, err := stub.InvokeRpc(reqCtx, method, req)
	if err != nil {
		return errors.WithStack(err)
	}

	return r.printMsg(w, res.(*dynamic.Message))
}

// setArgs sets argument values.
func (r ServerReflector) setArgs(req *dynamic.Message, descs []*desc.FieldDescriptor, args []string) error {
	for i, d := range descs {
		reflector := r.findArgReflector(d)
		if reflector == nil {
			// Not possible.
			return ErrUnsupportedReflectType
		}

		v, err := reflector.Parse(d, args[i])
		if err != nil {
			return err
		}

		req.SetFieldByName(d.GetName(), v)
	}

	return nil
}

// setFlags sets flag values.
func (r ServerReflector) setFlags(req *dynamic.Message, descs []*desc.FieldDescriptor, flags *pflag.FlagSet) error {
	for _, d := range descs {
		reflector := r.findFlagReflector(d)
		if reflector == nil {
			// Not possible.
			return ErrUnsupportedReflectType
		}

		v, err := reflector.ParseFlag(d, flags)
		if err != nil {
			return err
		}

		req.SetFieldByName(d.GetName(), v)
	}

	return nil
}

// printMsg prints a message received from the server.
func (r ServerReflector) printMsg(w io.Writer, msg *dynamic.Message) error {
	descs := msg.GetKnownFields()

	if len(descs) == 1 {
		value, err := r.pretty(descs[0], msg.GetField(descs[0]))
		if err != nil {
			return err
		}

		fmt.Println(value)
		return nil
	}

	tw := new(tabwriter.Writer)
	tw.Init(w, 0, 8, 2, ' ', 0)

	for _, d := range descs {
		label := strings.ToUpper(ReflectFieldDesc(d))

		value, err := r.pretty(d, msg.GetField(d))
		if err != nil {
			return err
		}

		fmt.Fprintf(tw, "%s\t%v\n", label, value)
	}

	return errors.WithStack(tw.Flush())
}

// printDynamicStream prints the messages of a server stream.
func (r ServerReflector) printStream(w io.Writer, ss *grpcdynamic.ServerStream, truncate, borders bool) error {
	var b bytes.Buffer
	var numCols int

	tw := new(tabwriter.Writer)
	tw.Init(&b, 0, 8, 1, ' ', 0)

	// Used to check if it should print the table header.
	first := true

	for {
		res, err := ss.RecvMsg()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.WithStack(err)
		}

		msg := res.(*dynamic.Message)
		descs := msg.GetKnownFields()
		numCols = len(descs)

		if first && numCols > 1 {
			r.printTableHeader(tw, descs, borders)
		}

		if err := r.printTableRow(tw, descs, msg, borders); err != nil {
			return err
		}

		first = false

	}

	if err := tw.Flush(); err != nil {
		return errors.WithStack(err)
	}

	// Truncate output.
	for i, row := range strings.Split(b.String(), "\n") {
		if row == "" {
			continue
		}

		if borders && numCols > 1 && i == 1 {
			// Print the horizontal line.
			row = strings.Replace(row, " ", "-", r.width)

			l := len(row) - 1
			if l > r.width {
				l = r.width
			}

			row = row[:l]
			fmt.Fprintln(w, row)

			continue
		}

		if truncate {
			r.printTruncated(w, row)
		} else {
			fmt.Fprintf(w, row)
		}
	}

	return nil
}

// printTableHeader prints the header of a table.
func (r ServerReflector) printTableHeader(w io.Writer, descs []*desc.FieldDescriptor, borders bool) {
	last := len(descs) - 1

	for i, d := range descs {
		label := strings.ToUpper(ReflectFieldDesc(d))

		if i < last {
			if borders {
				fmt.Fprintf(w, "%s\t| ", label)
			} else {
				fmt.Fprintf(w, "%s\t", label)
			}
		} else {
			fmt.Fprintf(w, "%s", label)
		}
	}

	fmt.Fprintln(w, "")

	// Print a dummy row for the horizontal line.
	if borders {
		for range descs[:last] {
			fmt.Fprint(w, "\t+")
		}

		fmt.Fprintln(w, "\t")
	}
}

// printTableRow prints a row of a table.
func (r ServerReflector) printTableRow(w io.Writer, descs []*desc.FieldDescriptor, msg *dynamic.Message, borders bool) error {
	last := len(descs) - 1

	for i, d := range descs {
		value, err := r.pretty(d, msg.GetFieldByName(d.GetName()))
		if err != nil {
			return err
		}

		if borders && i < last {
			fmt.Fprintf(w, "%s\t| ", value)
		} else {
			fmt.Fprintf(w, "%s\t", value)
		}
	}

	fmt.Fprintln(w, "")

	return nil
}

// printTruncated prints a strings and truncates it if its length is larger
// than the terminal width.
func (r ServerReflector) printTruncated(w io.Writer, s string) {
	s = strings.TrimRight(s, " ")
	runes := []rune(s)

	if r.width > 0 && len(runes) > r.width {
		fmt.Fprintln(w, string(runes[:r.width-3])+"...")
		return
	}

	fmt.Fprintln(w, s)
}

// pretty prints a human friendly string of a value.
func (r ServerReflector) pretty(d *desc.FieldDescriptor, v interface{}) (string, error) {
	reflector := r.findResReflector(d)
	if reflector == nil {
		return "", ErrUnsupportedReflectType
	}

	s, err := reflector.Pretty(d, v)
	if err != nil {
		return "", err
	}

	return s, err
}
