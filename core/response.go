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
	respWelcome      = &response{code: 220, info: "kll FTP server ready"}
	respTempReceived = &response{code: 331, info: "input password to login"}
	respLoginSuccess = &response{code: 200, info: "login successfully"}

	respSyntaxError  = &response{code: 500, info: "syntax error, command unrecognized"}
	respProcessError = &response{code: 553, info: "an error occur in the server, requested action not taken"}
	respParamsError  = &response{code: 504, info: "command not implemented for that parameter"}
	respLoginError   = &response{code: 530, info: "user need login"}
	respAuthError    = &response{code: 530, info: "user no auth to execute this command"}
)
