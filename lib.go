package libapi

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type DicJson map[string]interface{}

/* ************************************************************************************* */
/* ************************************************************************************* */

func Success(ctx *gin.Context, dic DicJson) {
	respuesta := DicJson{
		"success": true,
		"msg":     "",
		"data":    dic,
	}

	ctx.JSON(http.StatusOK, respuesta)
}

func Error(ctx *gin.Context, msg string) {
	respuesta := DicJson{
		"success": false,
		"msg":     msg,
	}

	ctx.JSON(http.StatusOK, respuesta)
}

/* ************************************************************************************* */
/* ************************************************************************************* */

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

type ARespuestaJson struct {
	Success bool   `json:"success" `
	Msg     string `json:"msg" `
}

func FactoryARespuestaJson(success bool, msg string) ARespuestaJson {
	return ARespuestaJson{success, msg}
}
