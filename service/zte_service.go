package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	logutil "github.com/tanmolk/ZXHN-H298A-Manager/utils"
	requestutil "github.com/tanmolk/ZXHN-H298A-Manager/utils"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var (
	sessionRegex     = regexp.MustCompile(`"_sessionTOKEN", "(\d+)"`)
	tempSessionRegex = regexp.MustCompile(`_sessionTmpToken = "(.*)"`)
	loginErrorRegex  = regexp.MustCompile(`var login_err_msg = "((\\x[0-9a-f]{2})+)";`)
)

type ZteService struct {
	gateway     string
	username    string
	password    string
	httpClient  *http.Client
	currentIPV6 string
}

func NewZteService(gateway string, username string, password string, client *http.Client) *ZteService {
	return &ZteService{
		gateway:    gateway,
		username:   username,
		password:   password,
		httpClient: client,
	}
}

func (zte *ZteService) Login() error {
	defer logutil.RecoverHandler()

	logutil.Normal("Login Starting")

	//get sessionId for login
	logutil.Normal("Getting sessionId for login")
	content, err := zte.getLoginPageContent(zte.gateway, true, nil)
	//check login state
	err = zte.checkLoginState(content)

	//if not login
	if err != nil {
		logutil.Normal("Login expired")
		regexResult := sessionRegex.FindStringSubmatch(content)
		if len(regexResult) < 2 {
			logutil.Error("Get login session id fails!")
			return errors.New("get login session id fails")
		}
		sessionId := regexResult[1]

		//get token
		logutil.Normal("Getting Token for login")
		token, err := zte.fetchLoginToken()

		if err != nil {
			logutil.Error(err)
			return err
		}

		//login
		logutil.Normal("Login")
		pageContent, err := zte.getLoginPageContent(zte.gateway, false, map[string][]string{
			"_sessionTOKEN": {sessionId},
			"action":        {"login"},
			"Username":      {zte.username},
			"Password":      {zte.passwordHash(token)},
		})

		if err != nil {
			logutil.Error("Try to login fails!")
			return err
		}

		//check login state
		err = zte.checkLoginState(pageContent)

		if err != nil {
			return err
		}
	}

	return nil
}

func (zte *ZteService) checkLoginState(pageContent string) error {
	logutil.Normal("Check Login")

	if strings.Contains(pageContent, "Frm_Password") {
		return errors.New("Login expired")
	}

	regexResult := loginErrorRegex.FindStringSubmatch(pageContent)
	if len(regexResult) == 3 {
		originalError := regexResult[1]
		errorString := zte.decodeHex(originalError)
		return errors.New(errorString)
	}
	return nil
}

func (zte *ZteService) decodeHex(hexCode string) string {
	codeHex := strings.ReplaceAll(hexCode, "\\x", "")
	bytes, _ := hex.DecodeString(codeHex)
	return string(bytes)
}

func (zte *ZteService) getLoginPageContent(gateWay string, ifFirst bool, values map[string][]string) (string, error) {
	var resp *http.Response
	var err error

	if ifFirst {
		resp, err = zte.httpClient.Get(gateWay)
	} else {
		resp, err = zte.httpClient.PostForm(gateWay, values)
	}

	if err != nil {
		logutil.Error("Request login page fails!")
	}

	//close
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	//read content
	allByte, err := io.ReadAll(resp.Body)
	if err != nil {
		logutil.Error("Read content fail!")
		return "", err
	}

	content := string(allByte)

	return content, err
}

func (zte *ZteService) passwordHash(loginToken string) string {
	sum256 := sha256.Sum256([]byte(zte.password + loginToken))
	return fmt.Sprintf("%x", sum256)
}

/*
 * Part From https://github.com/cheahjs/hyperoptic_zte_exporter
 */
type loginTokenResponse struct {
	XMLName    xml.Name `xml:"ajax_response_xml_root"`
	LoginToken string   `xml:",chardata"`
}

func (zte *ZteService) fetchLoginToken() (string, error) {
	tokenResp, err := zte.httpClient.Get(
		fmt.Sprintf("%s/function_module/login_module/login_page/logintoken_lua.lua", zte.gateway))
	defer tokenResp.Body.Close()

	if err != nil {
		logutil.Error("Get token fails")
		return "", err
	}

	body, _ := io.ReadAll(tokenResp.Body)

	var xmlResp loginTokenResponse
	_ = xml.Unmarshal(body, &xmlResp)
	return xmlResp.LoginToken, nil
}

type IPV6Response struct {
	Ip string `json:"ip"`
}

func (zte *ZteService) getCurrentIPV6() string {
	logutil.Normal("Start getCurrentIPV6")
	resp, _ := zte.httpClient.Get("https://api64.ipify.org?format=json")
	content := requestutil.ReadContent(resp)

	if content == nil {
		return ""
	}

	logutil.Normal(string(content))
	ipv6Response := IPV6Response{}

	err := json.Unmarshal(content, &ipv6Response)

	if err != nil {
		logutil.Error(err)
		return ""
	}
	logutil.Normal("End getCurrentIPV6")
	return ipv6Response.Ip

}

func (zte *ZteService) submitUpdateIPV6Form(ipv6 string) bool {
	logutil.Normal("Start submitUpdateIPV6Form")

	//request the indispensable page and get temporary token
	logutil.Normal("Get temporary token for submitUpdateIPV6Form")
	resp, err := zte.httpClient.Get("http://192.168.1.1/getpage.lua?pid=123&nextpage=Internet_Security_SecFilter_t.lp&Menu3Location=0")
	if err != nil {
		logutil.Error(err)
		return false
	}
	content := string(requestutil.ReadContent(resp))
	regexResult := tempSessionRegex.FindStringSubmatch(content)
	tempSessionToken := regexResult[1]

	//submit form

	values := map[string][]string{
		"IF_ACTION":             {"Apply"},
		"Enable":                {"1"},
		"_InstID":               {"IGD.FWIPv6Filter1"},
		"Name":                  {"test-ipv6"},
		"FilterTarget":          {"1"},
		"Protocol":              {"4"},
		"SrcIp":                 {"::"},
		"SrcPrefixLen":          {"-1"},
		"DstPrefixLen":          {"128"},
		"DstIp":                 {ipv6},
		"INCViewName":           {"DEV.IP.IF4"},
		"OUTCViewName":          {"DEV.IP.IF1"},
		"Btn_cancel_IPv6Filter": {},
		"Btn_apply_IPv6Fil":     {},
		"_sessionTOKEN":         {zte.decodeHex(tempSessionToken)},
	}

	logutil.Normal("submitUpdateIPV6Form")
	resp, _ = zte.httpClient.PostForm("http://192.168.1.1/common_page/IPv6Filter_lua.lua", values)

	respObj := &UpdateIPV6FormResponse{}
	readContent := requestutil.ReadContent(resp)
	err = xml.Unmarshal(readContent, respObj)
	//logutil.Normal(respObj)
	//logutil.Normal(string(readContent))
	//logutil.Normal(ipv6)

	if err != nil {
		logutil.Error(err)
		return false
	}

	return respObj.Result == "SUCC"
}

type UpdateIPV6FormResponse struct {
	Result string `xml:"IF_ERRORTYPE"`
}

func (zte *ZteService) UpdateIPV6() bool {
	defer logutil.RecoverHandler()

	logutil.Normal("Start UpdateIPV6")

	ipv6 := zte.getCurrentIPV6()

	if zte.currentIPV6 == ipv6 || ipv6 == "" {
		logutil.Normal(fmt.Sprintf("Update fails| Pre: %s -----> Now: %s", zte.currentIPV6, ipv6))
		return false
	}

	if zte.submitUpdateIPV6Form(ipv6) {
		if ChangeDNSRecord(ipv6) {
			logutil.Normal(fmt.Sprintf("Update Successfully| Pre: %s -----> Now: %s", zte.currentIPV6, ipv6))
			zte.currentIPV6 = ipv6
			return true
		}
	}

	logutil.Error("Update fails")
	return false
}
