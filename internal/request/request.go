package request

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strconv"

	"github.com/erespereza/new-project/pkg/validation"
)

type FormRequest interface {
	Rules() map[string]validation.Validation // se debe implementar, proposito: retornar las reglas de validacion
	PrepareForValidation() error             // se debe implementar, Propósito: Modifica o normaliza los datos del request y añadir lógica adicional antes de validar.
	WithValidator() error                    // se debe implementar, Propósito: Permite añadir lógica adicional después de preparar el validador pero antes de que se realice la validación.
	ParseQuery(r *http.Request)              // no se bede implementar, ya esta implementada en el Request
	Validate(req *http.Request)              // no se bede implementar, ya esta implementada en el Request
}

// Implementación de FormRequest para un struct
type Request struct {
	Query map[string]any
}

// Toma los valores de la url y los parsea en un map
func (r *Request) ParseQuery(req *http.Request) {
	// Inicializar el mapa Query si no está inicializado
	if r.Query == nil {
		r.Query = make(map[string]any)
	}

	// Obtener los parámetros de la URL
	queryParams := req.URL.Query()

	// Iterar sobre los parámetros de la URL
	for key, values := range queryParams {
		// El valor puede ser un solo valor o una lista, tomo solo el primer valor
		value := values[0]

		// Intentar convertir el valor a diferentes tipos
		if intValue, err := strconv.Atoi(value); err == nil {
			// Es un int
			r.Query[key] = intValue
		} else if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			// Es un float
			r.Query[key] = floatValue
		} else if boolValue, err := strconv.ParseBool(value); err == nil {
			// Es un bool
			r.Query[key] = boolValue
		} else {
			// Es un string (por defecto)
			r.Query[key] = value
		}
	}
}

func (r *Request) Validate(request FormRequest, req *http.Request) error {

	// Usar reflect para validar que se trabaja con el tipo especifico y obtener el tipo de request y deserializar en el tipo real
	requestValue := reflect.ValueOf(request)
	if requestValue.Kind() != reflect.Ptr || requestValue.IsNil() {
		return errors.New("se espera un puntero al tipo que implementa FormRequest")
	}

	// Leer el cuerpo de la solicitud
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	defer req.Body.Close()

	// Deserializar el JSON en el struct
	if err := json.Unmarshal(body, request); err != nil {
		return err
	}

	// Preparar el request antes de validar
	if err := request.PrepareForValidation(); err != nil {
		return err
	}

	// Añadir lógica adicional después de preparar el validador
	if err := request.WithValidator(); err != nil {
		return err
	}

	// Validar el request con las reglas de validación
	if err := validation.Struct(request, request.Rules()); err != nil {
		return err
	}

	// Si no hay errores, parsear los parámetros de la URL
	r.ParseQuery(req)

	return nil
}
