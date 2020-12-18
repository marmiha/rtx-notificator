// Logical interfaces for the messaging service clients.
package msgservice

type SenderClient interface {
	Send(msg string) ([]string, []error)
}