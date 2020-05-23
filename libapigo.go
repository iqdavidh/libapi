package libapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

/* Mocks ************************************************ */

var (
	enabledMocks = false
	mocks        = make(map[string]*Mock)
)

type Mock struct {
	Url        string
	HttpMethod string
	StringResponse   []byte
	Err        error
}

/* ************************************************ */
type DicJson map[string]interface{}

type DicResp struct {
	Success bool    `json:"success"`
	Msg     string  `json:"msg"`
	Data    DicJson `json:"data"`
}

/* ************************************************************************************* */
/* ************************************************************************************* */
//Enviar respuesta 200 con cuento success,msg y data con el  map[string]interface{} que se proporcione
func RespuestaSuccess(ctx *gin.Context, dic DicJson) {
	respuesta := DicResp{
		true,
		"",
		dic,
	}

	ctx.JSON(http.StatusOK, respuesta)
}

//Enviar respuesta 200 con cuento success=FALSE,msg
func RespuestaError(ctx *gin.Context, msg string) {
	respuesta := DicResp{
		false,
		msg,
		nil,
	}

	ctx.JSON(http.StatusOK, respuesta)
}

/* ************************************************************************************* */
/* ************************************************************************************* */

//Obtener un query parame integer desde ejemplo de ?param1=valorInteger
func GetIntFromQP(ctx *gin.Context, paramName string, valorDefault uint32, isOpcional bool) (uint32, error) {
	dicQueryParams := ctx.Request.URL.Query()

	paramAsString, isExiste := dicQueryParams[paramName]

	if !isExiste {

		if !isOpcional {
			return 0, errors.New("no se encontro el parametro " + paramName)
		} else {
			return valorDefault, nil
		}

	}

	if len(paramAsString[0]) > 10 {
		return 0, errors.New("longitud de texto - formato incorrecto " + paramName + "(" + paramAsString[0] + ")")
	}

	id, err := strconv.ParseInt(paramAsString[0], 10, 64)

	if err != nil {
		return 0, errors.New("no es un int - formato incorrecto parametro " + paramName)
	}

	return uint32(id), err
}

func GetStringFromQP(ctx *gin.Context, paramName string, valorDefault string, isOpcional bool) (string, error) {
	dicQueryParams := ctx.Request.URL.Query()

	paramAsString, isExiste := dicQueryParams[paramName]

	if !isExiste {

		if !isOpcional {
			return "", errors.New("no se encontro el parametro " + paramName)
		} else {
			return valorDefault, nil
		}

	}

	return paramAsString[0], nil

}

func GetDataCleanFromQP(ctx *gin.Context, listaCamposAllow []string) DicJson {

	dicQueryParams := ctx.Request.URL.Query()
	dic := DicJson{}

	for _, paramName := range listaCamposAllow {
		paramAsString, isExiste := dicQueryParams[paramName]
		if isExiste {
			dic[paramName] = paramAsString[0]
		}

	}

	return dic

}

func GetIsAllCamposRequeridos(dicCampos DicJson, listaCamposReq []string) (bool, error) {

	msgError := ""
	for _, paramName := range listaCamposReq {

		_, isExiste := dicCampos[paramName]
		if !isExiste {
			if msgError != "" {
				msgError += ", "
			}
			msgError += paramName
		}

	}

	if msgError != "" {
		return false, errors.New("campos faltantes " + msgError)
	}
	return true, nil

}

func DecodeBodyResponse(Body *bytes.Buffer) (DicResp, error) {

	body, _ := ioutil.ReadAll(Body)

	if !json.Valid(body) {
		return DicResp{}, errors.New("json inválido \n" + fmt.Sprint(body))
	}
	var respuesta DicResp
	erorrJson := json.Unmarshal(body, &respuesta)
	if erorrJson != nil {
		return DicResp{}, errors.New("json inválido \n" + fmt.Sprint(body))
	}

	return respuesta, nil
}

/* *********************************************************************************************** */
/* *********************************************************************************************** */
/* API Clien  */


func getMockId(httpMethod string, url string) string {
	return fmt.Sprintf("%s_%s", httpMethod, url)
}

func APiClientStartMockups() {
	enabledMocks = true
}

func ApiClientFlushMockups() {
	mocks = make(map[string]*Mock)
}

func ApiClientStopMockups() {
	enabledMocks = false
}

func ApiClientAddMockup(mock Mock) {
	mocks[getMockId(mock.HttpMethod, mock.Url)] = &mock
}


func getResponseForRequestJSON(req *http.Request, dicHeader map[string]string) (DicJson, error) {

	req.Header.Set("Content-type", "application/json")

	for k := range dicHeader {
		req.Header.Set(k, dicHeader[k])
	}

	timeout := time.Duration(10 * time.Second)
	client := http.Client{Timeout: timeout}

	/* Hacer request********************* */
	res, errRespo := client.Do(req)
	if errRespo != nil {
		return DicJson{}, errRespo
	}
	/* ************************************ */

	defer res.Body.Close()

	jsonBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return DicJson{}, err
	}

	//convertir body en JSON
	dic := DicJson{}
	errJson := json.Unmarshal(jsonBody, &dic)

	if errJson != nil {
		return DicJson{}, errJson
	}

	return dic, nil
}

func ApiClientReqGETJson(url string, dicHeader map[string]string) (DicJson, error) {

	if enabledMocks {
		mock := mocks[getMockId(http.MethodGet, url)]
		if mock == nil {
			return nil, errors.New("no mockup found for give request")
		}

		var dic DicJson
		_ = json.Unmarshal(mock.StringResponse, &dic)
		
		return dic, mock.Err
	}

	req, errFac := http.NewRequest("GET", url, nil)

	if errFac != nil {
		return DicJson{}, errFac
	}

	return getResponseForRequestJSON(req, dicHeader)
}

func ApiClientReqPOSTJson(url string, dicHeader map[string]string, bodyJson []byte) (DicJson, error) {

	if enabledMocks {
		mock := mocks[getMockId(http.MethodPost, url)]
		if mock == nil {
			return nil, errors.New("no mockup found for give request")
		}

		var dic DicJson
		_ = json.Unmarshal(mock.StringResponse, &dic)

		return dic, mock.Err
	}

	
	req, errFac := http.NewRequest("POST", url, bytes.NewBuffer(bodyJson))

	if errFac != nil {
		return DicJson{}, errFac
	}

	return getResponseForRequestJSON(req, dicHeader)
}

/* *********************************************************************************************** */
/* *********************************************************************************************** */
/*TEsting*/

type ConfigTestBasic struct {
	CodeRespuesta   int
	DicHeader       map[string]string
	QueryParams     string
	UrlParamsValor  string
	UrlParamsPatron string
	Body string
}

func FactoryConfigTestBasic(dicHeader map[string]string) ConfigTestBasic {
	return ConfigTestBasic{
		CodeRespuesta: 200,
		DicHeader:     dicHeader,
	}
}
func TestBasicRequestGET(t2 *testing.T, a *assert.Assertions, fnHandler gin.HandlerFunc, configTest ConfigTestBasic) DicResp {
	gin.SetMode(gin.TestMode)

	route := gin.Default()

	urlMap, urlConQueryParams := configRouteBasicTest(configTest)

	route.GET(urlMap, fnHandler)
	req, errReq := http.NewRequest(http.MethodGet, urlConQueryParams, nil)
	if errReq != nil {
		fmt.Println(errReq)
		t2.Fatalf("Couldn't create request: %v\n", errReq)
	}

	for k := range configTest.DicHeader {
		req.Header.Set(k, configTest.DicHeader[k])
	}

	w := httptest.NewRecorder()
	route.ServeHTTP(w, req)

	a.True(w.Code == configTest.CodeRespuesta, "No es el codigo Esperado")

	respuesta, errorDecode := DecodeBodyResponse(w.Body)

	a.True(errorDecode == nil, "Esperamos error nil "+fmt.Sprint(errorDecode))

	if errorDecode != nil {
		t2.Fatalf(fmt.Sprint(errorDecode))
	}

	return respuesta

}

func configRouteBasicTest(configTest ConfigTestBasic) (string, string) {
	urlMap := "/mock"
	if configTest.UrlParamsPatron != "" {
		urlMap = urlMap + configTest.UrlParamsPatron
	}

	urlTest := "/mock"
	if configTest.UrlParamsValor != "" {
		urlTest = urlTest + configTest.UrlParamsValor
	}

	urlConQueryParams := urlTest
	if configTest.QueryParams != "" {
		urlConQueryParams = urlConQueryParams + "?" + configTest.QueryParams
	}
	return urlMap, urlConQueryParams
}

func TestBasicRequestPOST(t2 *testing.T, a *assert.Assertions, fnHandler gin.HandlerFunc, configTest ConfigTestBasic) DicResp {
	gin.SetMode(gin.TestMode)
	route := gin.Default()

	urlMap, urlConQueryParams := configRouteBasicTest(configTest)

	route.POST(urlMap, fnHandler)
	req, errReq := http.NewRequest(http.MethodPost, urlConQueryParams, strings.NewReader(configTest.Body))

	if errReq != nil {
		fmt.Println(errReq)
		t2.Fatalf("Couldn't create request: %v\n", errReq)
	}

	for k := range configTest.DicHeader {
		req.Header.Set(k, configTest.DicHeader[k])
	}

	w := httptest.NewRecorder()
	route.ServeHTTP(w, req)

	a.True(w.Code == configTest.CodeRespuesta, "No es el codigo Esperado")

	respuesta, errorDecode := DecodeBodyResponse(w.Body)

	a.True(errorDecode == nil, "Esperamos error nil "+fmt.Sprint(errorDecode))

	if errorDecode != nil {
		t2.Fatalf(fmt.Sprint(errorDecode))
	}

	return respuesta

}
