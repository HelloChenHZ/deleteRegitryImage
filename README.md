# deleteRegistryImage
golang delete old image layers in registry v2

![](./logo.jpg)

# Usage
```
Usage of ./deleteRegistryImage:
  -keepNumber int
    	the digest numbers keep in each tag (default 10)
  -projectName string
    	project name (default "infra/docker/codis/")
  -registryDir string
    	registry directory (default "/gitlab-registry/docker/registry/v2/")

```

# TODO
- refactor the code struct(add struct.go function.go)
- path judgement
- multi threads