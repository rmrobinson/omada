# api

This package contains an auto-generated API implementation of the Omada OpenAPI specification. I used the modified openapi.json file referenced at https://github.com/charlie-haley/omada_exporter/issues/81 since, as the GH issue calls out, the Omada JSON file returned by the SDN controller does not parse. I had to manually convert the `*/*` response type qualifier in the modified openapi.json file to `application/json` to hint to the oapi-codegen tool that it needed to create the JSON200 field in the result structs.

After installing the [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) tool, the following command was used:

```
oapi-codegen --config oapi-config.yaml openapi.json
```

TODO: convert to use the go-generate framework