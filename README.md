## Build and run using docker (windows)

Build the docker image.
```
docker build . -t deploy-agentv1 -f ./build/deploy-agent/Dockerfile
```
Run it...
```
docker run `
--name dtest `
--mount type=bind,source="$(pwd)"/config-example.json,target=/config-example.json `
deploy-agentv1 --config config-example.json
```