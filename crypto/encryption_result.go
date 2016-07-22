package crypto

type encryptionResult struct {
	key     string
	message EncryptedMessage
	err     error
}

func (er encryptionResult) Key() string {
	return er.key
}

func (er encryptionResult) Message() EncryptedMessage {
	return er.message
}

func (er encryptionResult) IsErr() bool {
	return er.err != nil
}

func (er encryptionResult) Error() string {
	if !er.IsErr() {
		return ""
	}
	return er.err.Error()
}

type EncryptionResult interface {
	Key() string
	Message() EncryptedMessage
	error
}

type encryptionResults []encryptionResult

func (er *encryptionResults) Add(r encryptionResult) {
	*er = append(*er, r)
}

func (er encryptionResults) Size() int {
	return len(er)
}

func (er encryptionResults) IsErr() bool {
	if er == nil {
		return false
	}
	// If even one encryption was a success, we don't consider this an error
	for _, v := range er {
		if v.err == nil {
			return false
		}
	}
	return true
}

func (er encryptionResults) Error() string {
	if er == nil {
		return ""
	}
	// Return first error
	for _, v := range er {
		if v.err != nil {
			return v.err.Error()
		}
	}
	return ""
}
