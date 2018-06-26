# Clock Service

The Clock Service is a sample service that can be used as a template for
new services.

It exposes two gRPC APIs:

- `clock-local` that returns the current time
- `clock-remote` that connects to a remote node to ask its current time
