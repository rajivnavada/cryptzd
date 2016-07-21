package crypto

type encryptionResult struct {
	key     string
	message EncryptedMessage
	err     error
}

type encryptionResults []encryptionResult

func (er encryptionResults) Add(r encryptionResult) {
	er = append(er, r)
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
