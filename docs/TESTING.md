# Testing

`go test ./...` runs everywhere. A few checks only mean something in a container run as one user, with folders that belong to another, as WittyMediaServer runs on a NAS. They skip elsewhere.

## The user and group note, in Docker

What the browser says when a folder can't be read in a container (issue #2) needs a folder the container's user can't open, and a mount point with nothing in it. On Linux or WSL with Docker, from this folder:

```sh
# The tests, built here and run as user and group 999:10, with nothing the container needn't have.
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go test -c -o /tmp/wifb.test .
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /tmp/wifb-helper ./cmd/wittyintegratedfilebrowser-module

# A folder that belongs to someone else, with a folder inside it, as media on a NAS often does.
sudo rm -rf /tmp/wifb-foreign && sudo mkdir -p /tmp/wifb-foreign/inner && sudo chown -R 1000:100 /tmp/wifb-foreign && sudo chmod 700 /tmp/wifb-foreign

docker run --rm --user 999:10 --read-only --cap-drop ALL --security-opt no-new-privileges \
  -v /tmp/wifb.test:/t/wifb.test:ro -v /tmp/wifb-helper:/t/helper:ro \
  -v /tmp/wifb-foreign:/media/foreign:ro --tmpfs /media/empty --tmpfs /tmp \
  -e WIFB_TEST_FOREIGN=/media/foreign -e WIFB_TEST_EMPTY_MOUNT=/media/empty \
  debian:stable-slim /t/wifb.test -test.v -test.run 'Explain|List'
```

`TestExplainInDocker` passes and logs `denied: runs as 999:10, folder belongs to 1000:100`. The helper itself, started with as little environment as a server gives it:

```sh
docker run --rm -i --user 999:10 --read-only --cap-drop ALL \
  -v /tmp/wifb-helper:/t/helper:ro -v /tmp/wifb-foreign:/media/foreign:ro --tmpfs /media/empty \
  debian:stable-slim env -i /t/helper --wms-helper <<'EOF'
{"id":1,"method":"list","params":{"path":"/media/foreign"}}
{"id":2,"method":"access","params":{"path":"/media/foreign"}}
{"id":3,"method":"access","params":{"path":"/media/foreign/inner"}}
{"id":4,"method":"access","params":{"path":"/media/empty"}}
{"id":5,"method":"access","params":{"path":"/media"}}
EOF
```

It answers `denied` for 1, a note naming `/media/foreign` for 2 and 3, `empty-mount` for 4 and `null` for 5. Run outside a container, every `access` is `null`.
