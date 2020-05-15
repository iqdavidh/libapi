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
//Enviar respuesta 200 con cuento success,msg y data con el dicJson que se proporcione
func RespuestaSuccess(ctx *gin.Context, dic DicJson) {
	respuesta := DicJson{
		"success": true,
		"msg":     "",
		"data":    dic,
	}

	ctx.JSON(http.StatusOK, respuesta)
}

//Enviar respuesta 200 con cuento success=FALSE,msg
func RespuestaError(ctx *gin.Context, msg string) {
	respuesta := DicJson{
		"success": false,
		"msg":     msg,
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
