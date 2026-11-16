package dtos

type EncryptedRegisterRequest struct {
	Username               string `json:"username" binding:"required,min=3,max=32"`
	Email                  string `json:"email" binding:"required,email"`
	AuthKey                string `json:"auth_key" binding:"required"`
	KDFAlgorithm           string `json:"kdf_algorithm" binding:"required"`
	KDFIterations          int    `json:"kdf_iterations" binding:"required,min=100000"`
	KDFVersion             int    `json:"kdf_version" binding:"required,min=1"`
	Salt                   string `json:"salt" binding:"required"`
	WrappedDataKey         string `json:"wrapped_data_key" binding:"required"`
	WrappedKeyIV           string `json:"wrapped_data_key_iv" binding:"required"`
	RecoveryWrappedDataKey string `json:"recovery_wrapped_data_key" binding:"required"`
	RecoveryWrappedKeyIV   string `json:"recovery_wrapped_data_key_iv" binding:"required"`
	RecoveryVerifier       string `json:"recovery_verifier" binding:"required"`
}

type LoginRequest struct {
	Email   string `json:"email" binding:"required,email"`
	AuthKey string `json:"auth_key" binding:"required"`
}

type LegacyLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LegacySetupRequest struct {
	AuthKey                string `json:"auth_key" binding:"required"`
	KDFAlgorithm           string `json:"kdf_algorithm" binding:"required"`
	KDFIterations          int    `json:"kdf_iterations" binding:"required,min=100000"`
	KDFVersion             int    `json:"kdf_version" binding:"required,min=1"`
	Salt                   string `json:"salt" binding:"required"`
	WrappedDataKey         string `json:"wrapped_data_key" binding:"required"`
	WrappedKeyIV           string `json:"wrapped_data_key_iv" binding:"required"`
	RecoveryWrappedDataKey string `json:"recovery_wrapped_data_key" binding:"required"`
	RecoveryWrappedKeyIV   string `json:"recovery_wrapped_data_key_iv" binding:"required"`
	RecoveryVerifier       string `json:"recovery_verifier" binding:"required"`
}

type ChangePasswordRequest struct {
	CurrentAuthKey string `json:"current_auth_key" binding:"required"`
	AuthKey        string `json:"auth_key" binding:"required"`
	KDFAlgorithm   string `json:"kdf_algorithm" binding:"required"`
	KDFIterations  int    `json:"kdf_iterations" binding:"required,min=100000"`
	KDFVersion     int    `json:"kdf_version" binding:"required,min=1"`
	Salt           string `json:"salt" binding:"required"`
	WrappedDataKey string `json:"wrapped_data_key" binding:"required"`
	WrappedKeyIV   string `json:"wrapped_data_key_iv" binding:"required"`
}

type RecoveryRequest struct {
	Email            string `json:"email" binding:"required,email"`
	RecoveryVerifier string `json:"recovery_verifier" binding:"required"`
	AuthKey          string `json:"auth_key" binding:"required"`
	KDFAlgorithm     string `json:"kdf_algorithm" binding:"required"`
	KDFIterations    int    `json:"kdf_iterations" binding:"required,min=100000"`
	KDFVersion       int    `json:"kdf_version" binding:"required,min=1"`
	Salt             string `json:"salt" binding:"required"`
	WrappedDataKey   string `json:"wrapped_data_key" binding:"required"`
	WrappedKeyIV     string `json:"wrapped_data_key_iv" binding:"required"`
}

type LegacyImportRequest struct {
	Questions []LegacyEncryptedQuestion `json:"questions"`
	Khayals   []LegacyEncryptedKhayal   `json:"khayals"`
}

type LegacyEncryptedQuestion struct {
	ID                int    `json:"id" binding:"required"`
	ClientID          string `json:"client_id" binding:"required"`
	Ciphertext        string `json:"ciphertext" binding:"required"`
	EncryptionVersion int    `json:"encryption_version" binding:"required"`
}

type LegacyEncryptedKhayal struct {
	ID                int    `json:"id" binding:"required"`
	QuestionID        int    `json:"question_id" binding:"required"`
	ClientID          string `json:"client_id" binding:"required"`
	Ciphertext        string `json:"ciphertext" binding:"required"`
	EncryptionVersion int    `json:"encryption_version" binding:"required"`
}
