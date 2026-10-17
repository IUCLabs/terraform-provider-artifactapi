package provider

func init() { registerResource(newRemoteResource("helm")) }
