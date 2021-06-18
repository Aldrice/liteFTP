package core

type response struct {
	code int
	info string
}

func createResponse(c int, i string) *response {
	return &response{
		code: c,
		info: i,
	}
}

var (
	respWelcome      = &response{code: 220, info: "KLL FTP server ready."}
	respTempReceived = &response{code: 331, info: "Input password to login."}
	respLoginSuccess = &response{code: 200, info: "Login successfully."}

	respSyntaxError  = &response{code: 500, info: "Syntax error, command unrecognized."}
	respProcessError = &response{code: 553, info: "An error occur in the server, requested action not taken."}
	respParamsError  = &response{code: 504, info: "Command not implemented for that parameter."}
	respAuthError    = &response{code: 530, info: "User no auth to execute this command."}
)
