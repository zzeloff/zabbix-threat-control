package zabbix

import "context"

// NewV7 constructs and authenticates a Zabbix 7.x (also 6.4+) client. The
// session/API token is presented via the Authorization: Bearer header and
// user.login uses the "username" parameter.
func NewV7(ctx context.Context, opts Options) (Client, error) {
	c := newRPCClient(opts, false)
	if opts.Token != "" {
		c.auth = opts.Token
		return c, nil
	}
	if err := c.login(ctx, "username", opts.User, opts.Password); err != nil {
		return nil, err
	}
	return c, nil
}
