# Storage

Object storage contract (`github.com/dtalabs/go-packages/storage`) and an S3-compatible
implementation (`github.com/dtalabs/go-packages/storage/adapters`). Supports object keys, signed
upload/download URLs, existence checks, and deletion.

```go
store, err := adapters.New(ctx, "s3://my-bucket/littledays", "ap-southeast-1")
if err != nil {
    return err
}
key := store.ObjectKey(userID, uploadID, "photo.jpg")
uploadURL, err := store.PresignPut(ctx, key, "image/jpeg")
```

Use `s3://bucket/prefix` for AWS or `http://endpoint/bucket/prefix` /
`https://endpoint/bucket/prefix` for S3-compatible services. Custom endpoints use
path-style bucket addressing. For local RustFS, use
`http://localhost:9000/bifrost` with region `us-east-1`. Credentials use the AWS SDK's default configuration.
For Cloudflare R2, use region `auto`.

Upload URLs expire after 15 minutes; download URLs expire after one hour. An
empty storage URL disables storage: `Configured()` returns false, signing and
existence checks return `storage.ErrNotConfigured`, and deletion is a no-op.

Validate from this directory:

```bash
go vet ./...
go build ./...
```
