# Publishing the Go SDK

Go doesn't have a registry — publishing = pushing to a public Git repo with versioned tags.

## One-time setup

1. **Create a GitHub repo** named e.g. `kriya-go` under your account or org.
2. **Edit `go.mod`** — replace the module path with your repo's URL:
   ```diff
   - module github.com/omkarjaliparthi/kriya-go
   + module github.com/<your-org>/<your-repo-name>
   ```
3. **Edit `client.go`** package doc comment + import examples if you rename the package.
4. **Push**:
   ```bash
   cd sdk/go
   git init
   git add .
   git commit -m "Initial release"
   git branch -M main
   git remote add origin git@github.com:omkarjaliparthi/kriya-go.git
   git push -u origin main
   git tag v0.1.0
   git push --tags
   ```

## Subsequent releases

```bash
# Bump the tag, push it.
git tag v0.1.1
git push --tags
```

Go modules pick up tags automatically; users get the new version with `go get …@latest` or `go get …@v0.1.1`.

## Consumers

```bash
go get github.com/omkarjaliparthi/kriya-go@latest
```
