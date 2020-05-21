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
/* Request  */

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

func RequestGETJson(url string, dicHeader map[string]string) (DicJson, error) {

	req, errFac := http.NewRequest("GET", url, nil)

	if errFac != nil {
		return DicJson{}, errFac
	}

	return getResponseForRequestJSON(req, dicHeader)
}

func RequestPOSTJson(url string, dicHeader map[string]string, bodyJson []byte) (DicJson, error) {

	req, errFac := http.NewRequest("POST", url, bytes.NewBuffer(bodyJson))

	if errFac != nil {
		return DicJson{}, errFac
	}

	return getResponseForRequestJSON(req, dicHeader)
}

/* *********************************************************************************************** */
/* *********************************************************************************************** */
/*TEsting*/

func TestBasicRequestGET(t2 *testing.T, a *assert.Assertions, queryParams string, group gin.HandlerFunc, codeRespuesta int, dicHeader map[string]string) DicResp {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	url := "/test"
	r.GET(url, group)

	urlConQueryParams := url
	if queryParams != "" {
		urlConQueryParams = urlConQueryParams + "?"+queryParams
	}

	req, errReq := http.NewRequest(http.MethodGet, urlConQueryParams, nil)
	if errReq != nil {
		fmt.Println(errReq)
		t2.Fatalf("Couldn't create request: %v\n", errReq)
	}

	for k := range dicHeader {
		req.Header.Set(k, dicHeader[k])
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	a.True(w.Code == codeRespuesta, "No es el codigo Esperado")

	respuesta, errorDecode := DecodeBodyResponse(w.Body)

	a.True(errorDecode == nil, "Esperamos error nil "+fmt.Sprint(errorDecode))

	if errorDecode != nil {
		t2.Fatalf(fmt.Sprint(errorDecode))
	}

	return respuesta

}

func TestBasicRequestPOST(t2 *testing.T, a *assert.Assertions, queryParams string, body string, handlerRequest gin.HandlerFunc, codeRespuesta int, dicHeader map[string]string) DicResp {
	gin.SetMode(gin.TestMode)
	ro := gin.Default()
	url := "/test"
	ro.POST(url, handlerRequest)

	urlConQueryParams := url
	if queryParams != "" {
		urlConQueryParams = urlConQueryParams + "?"+queryParams
	}

	req, errReq := http.NewRequest(http.MethodPost, urlConQueryParams, strings.NewReader(body))
	if errReq != nil {
		fmt.Println(errReq)
		t2.Fatalf("Couldn't create request: %v\n", errReq)
	}

	for k := range dicHeader {
		req.Header.Set(k, dicHeader[k])
	}

	w := httptest.NewRecorder()
	ro.ServeHTTP(w, req)

	a.True(w.Code == codeRespuesta, "No es el codigo Esperado")

	respuesta, errorDecode := DecodeBodyResponse(w.Body)

	a.True(errorDecode == nil, "Esperamos error nil "+fmt.Sprint(errorDecode))

	if errorDecode != nil {
		t2.Fatalf(fmt.Sprint(errorDecode))
	}

	return respuesta

}
