# Omada API Client

Newer versions of the Omada SDN controller allow for programmatic access to functionality through an API. The specification for this API is defined in using [OpenAPI](https://use1-omada-northbound.tplinkcloud.com/doc.html#/home). While the majority of functionality is specified using OpenAPI, the authentication framework is implemented separately.

This project is a thin wrapper around the auto-generated OpenAPI classes that includes authentication token handling; automatically obtaining an access and refresh token using the supplied information. This assumes the client credential flow is used.

See an example in `cmd/omadacli` for a small tool which finds all the sites that the controller is managing and dumps all the wireless devices connected.
