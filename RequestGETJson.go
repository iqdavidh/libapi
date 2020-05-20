package libapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strconv"
)

type DicJson map[string]interface{}

type DicResp struct {
	Success bool `json:"success"`
	Msg     string `json:"msg"`
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

/* *********************************************************************************************** */
/* *********************************************************************************************** */
/* *********************************************************************************************** */

func RequestGETJson(url string, headers map[string]string) (DicJson, error) {

	res, err := http.Get(url)
	if err != nil {
		return DicJson{}, err
	}
	defer res.Body.Close()

	jsonBody, err2 := ioutil.ReadAll(res.Body)
	if err2 != nil {
		return DicJson{}, err2
	}

	//convertir body en JSON
	dic := DicJson{}
	errJson := json.Unmarshal(jsonBody, &dic)

	if errJson != nil {
		return DicJson{}, errJson
	}

	return dic, nil
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
