package kubejson

import (
	"github.com/justinsb/kweb/components/kube/kubejson/internal/encoding/json"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"k8s.io/klog/v2"
)

func (o UnmarshalOptions) UnmarshalKubeList(b []byte, m proto.Message, callback func(m proto.Message)) (*KubeListObjectMeta, error) {
	return o.unmarshalKubeList(b, m, callback)
}

// unmarshal is a centralized function that all unmarshal operations go through.
// For profiling purposes, avoid changing the name of this function or
// introducing other code paths for unmarshal that do not go through this.
func (o UnmarshalOptions) unmarshalKubeList(b []byte, m proto.Message, callback func(m proto.Message)) (*KubeListObjectMeta, error) {
	// proto.Reset(m)

	if o.Resolver == nil {
		o.Resolver = protoregistry.GlobalTypes
	}

	dec := decoder{json.NewDecoder(b), o}
	meta, err := dec.unmarshalKubeListObject(m.ProtoReflect(), callback)
	if err != nil {
		return nil, err
	}

	// Check for EOF.
	tok, err := dec.Read()
	if err != nil {
		return nil, err
	}
	if tok.Kind() != json.EOF {
		return nil, dec.unexpectedTokenError(tok)
	}

	return meta, nil
}

type KubeListObjectMeta struct {
	APIVersion string
	Kind       string
}

// unmarshalMessage unmarshals a message into the given protoreflect.Message.
func (d decoder) unmarshalKubeListObject(m protoreflect.Message, callback func(m proto.Message)) (*KubeListObjectMeta, error) {
	meta := &KubeListObjectMeta{}

	tok, err := d.Read()
	if err != nil {
		return nil, err
	}
	if tok.Kind() != json.ObjectOpen {
		return nil, d.unexpectedTokenError(tok)
	}

	for {
		// Read field name.
		tok, err := d.Read()
		if err != nil {
			return nil, err
		}
		switch tok.Kind() {
		default:
			return nil, d.unexpectedTokenError(tok)
		case json.ObjectClose:
			return meta, nil
		case json.Name:
			// Continue below.
		}

		name := tok.Name()

		switch name {
		case "apiVersion":
			apiVersion, err := d.readString()
			if err != nil {
				return nil, err
			}
			meta.APIVersion = apiVersion

		case "kind":
			kind, err := d.readString()
			if err != nil {
				return nil, err
			}
			meta.Kind = kind

		case "metadata":
			klog.Warningf("skipping metadata on list response")
			if err := d.skipJSONValue(); err != nil {
				return nil, err
			}

		case "items":
			tok, err := d.Read()
			if err != nil {
				return nil, err
			}
			if tok.Kind() != json.ArrayOpen {
				return nil, d.unexpectedTokenError(tok)
			}

			for {
				// Check for End of slice.
				tok, err := d.Peek()
				if err != nil {
					return nil, err
				}
				if tok.Kind() == json.ArrayClose {
					_, err := d.Read()
					if err != nil {
						return nil, err
					}
					break
				}

				elem := m.New()
				if err := d.unmarshalMessage(elem, false); err != nil {
					return nil, err
				}

				// if !p.AllowPartial {
				// 	if err := proto.CheckInitialized(m); err != nil {
				// 		return nil, err
				// 	}
				// }

				callback(elem.Interface())
			}

		default:
			// Field is unknown.
			// if d.opts.DiscardUnknown {
			// 	if err := d.skipJSONValue(); err != nil {
			// 		return err
			// 	}
			// 	continue
			// }
			return nil, d.newError(tok.Pos(), "unknown field %v", tok.RawString())

		}

		// // TODO: Handle these fields properly (there is google.protobuf.Timestamp)
		// if name == "creationTimestamp" || name == "generation" || name == "managedFields" {
		// 	if err := d.skipJSONValue(); err != nil {
		// 		return err
		// 	}
		// 	continue
		// }
	}
}

func (d decoder) readString() (string, error) {
	tok, err := d.Read()
	if err != nil {
		return "", err
	}

	if tok.Kind() == json.String {
		return tok.ParsedString(), nil
	}

	return "", d.newError(tok.Pos(), "invalid value for string type: %v", tok.RawString())
}
