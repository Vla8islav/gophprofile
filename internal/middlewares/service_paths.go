package middlewares

var ServicePaths = map[string]bool{
	"/health":  true,
	"/ready":   true,
	"/metrics": true,
}
