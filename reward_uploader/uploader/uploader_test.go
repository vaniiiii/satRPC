package uploader

import (
	"context"
	"testing"
)

func testUploader(t *testing.T) {
	ctx := context.Background()
	operators := "bbn1rt6v30zxvhtwet040xpdnhz4pqt8p2za7y430x&bbn1nrueqkp0wmujyxuqp952j8mnxngm5gek3fsgrj"

	u := NewUploader()
	u.calcReward(ctx, 1, "1", operators)
}
