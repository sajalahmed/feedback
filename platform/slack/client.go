package slack

import "log"

type Client interface {
	PostMessage(channel string, message string) error
}

type MockClient struct {
}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) PostMessage(channel string, message string) error {
	log.Printf("[Mock Slack] Posting to %s: %s", channel, message)
	return nil
}

type RealClient struct {
	Token string
}

func (c *RealClient) PostMessage(channel string, message string) error {
	return nil
}
