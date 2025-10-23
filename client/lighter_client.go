package client

// LighterWebsocketClient is the main entry point, similar to bybit.NewWebsocketClient()
type LighterWebsocketClient struct {
	config *WSConfig
}

// NewLighterWebsocketClient creates a new Lighter WebSocket client
func NewLighterWebsocketClient() *LighterWebsocketClient {
	return &LighterWebsocketClient{
		config: DefaultWSConfig(),
	}
}

// SetConfig allows customizing the WebSocket configuration
func (c *LighterWebsocketClient) SetConfig(config *WSConfig) *LighterWebsocketClient {
	c.config = config
	return c
}

// Public returns the public market data service
func (c *LighterWebsocketClient) Public() (LighterWebsocketPublicServiceI, error) {
	service := NewLighterWebsocketPublicService(c.config)
	return service, nil
}

// Private returns the private account data service
func (c *LighterWebsocketClient) Private(tokenGen TokenGenerator) (LighterWebsocketPrivateServiceI, error) {
	service := NewLighterWebsocketPrivateService(c.config, tokenGen)
	return service, nil
}