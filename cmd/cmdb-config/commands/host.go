package commands

import "github.com/urfave/cli"

func getHost(c *cli.Context) string {
	host := c.String("host")
	// TODO - Validation of IP and Port
	if len(host) == 0 {
		return "localhost:9090"
	}
	return host
}

func getApiKey(c *cli.Context) string {
	apikey := c.String("apikey")
	if len(apikey) == 0 {
		return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBY2NvdW50SUQiOjF9.GsXyFDDARjXe1t9DPo2LIBKHEal3O7t3vLI3edA7dGU"
	}
	return apikey
}
