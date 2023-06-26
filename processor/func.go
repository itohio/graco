package processor

import (
	"context"
	"encoding/json"
	"errors"
)

func MarshalJSON[T any]() processFuncWrapper[T, Blob[T]] {
	return Func[T, Blob[T]](
		func(ctx context.Context, val T) (Blob[T], error) {
			bytes, err := json.Marshal(val)
			return NewBlob[T](val, bytes), err
		},
	)
}

func UnmarshalJSON[T any]() processFuncWrapper[Blob[T], T] {
	return Func[Blob[T], T](
		func(ctx context.Context, blob Blob[T]) (T, error) {
			var val T
			err := json.Unmarshal(blob.Data(), &val)
			return val, errors.Join(err, blob.Close())
		},
	)
}
