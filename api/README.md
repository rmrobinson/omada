# api

This package contains an auto-generated API implementation of the Omada OpenAPI specification. I used the OpenAPI JSON file retrieved from my Omada Controller running software version 5.14.26.1 at https://<hostname>/v3/api-docs, with several modifications:
- I converted the `*/*` response type qualifier in the schema to `application/json` to hint to the oapi-codegen tool that it should create the JSON200 field in the result structs
- I removed the `wanList` property in the `CheckWanLanStatusOpenApiVO` object since it referenced a non-existant result type
- I updated 21 API methods which had missing path parameter definitions

If you update the openapi.json file, you can regenerate the API implementation by running `go generate` in this directory.

NOTE: newer versions of the OpenAPI specification retrieved from the Omada controller may require similar, or additional, modifications by hand before the generate step will succeed.