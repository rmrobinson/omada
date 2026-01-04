# api

This package contains an auto-generated API implementation of the Omada OpenAPI specification. I used the OpenAPI JSON file retrieved from my Omada Controller running software version 5.14.26.1 at https://<hostname>/v3/api-docs, with several modifications:
- I converted the `*/*` response type qualifier in the JSON file file to `application/json` to hint to the oapi-codegen tool that it needed to create the JSON200 field in the result structs
- I removed the `wanList` property in the `CheckWanLanStatusOpenApiVO` object since it referenced a non-existant result type
- I updated 21 API methods which had missing path parameter definitions

After installing the [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) tool, the following command was used:

```
oapi-codegen --config oapi-config.yaml openapi.json
```

TODO: convert to use the go-generate framework