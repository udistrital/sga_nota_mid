package controllers

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/sga_mid_notas/services"
	"github.com/udistrital/utils_oas/errorhandler"
	"github.com/udistrital/utils_oas/requestresponse"
)

// NotasController operations for Notas
type NotasController struct {
	beego.Controller
}

// URLMapping ...
func (c *NotasController) URLMapping() {
	c.Mapping("GetEspaciosAcademicosDocente", c.GetEspaciosAcademicosDocente)
	c.Mapping("GetModificacionExtemporanea", c.GetModificacionExtemporanea)
	c.Mapping("GetDatosDocenteAsignatura", c.GetDatosDocenteAsignatura)
	c.Mapping("GetPorcentajesAsignatura", c.GetPorcentajesAsignatura)
	c.Mapping("PutPorcentajesAsignatura", c.PutPorcentajesAsignatura)
	c.Mapping("GetCapturaNotas", c.GetCapturaNotas)
	c.Mapping("PutCapturaNotas", c.PutCapturaNotas)
	c.Mapping("GetEstadosRegistros", c.GetEstadosRegistros)
	c.Mapping("GetDatosEstudianteNotas", c.GetDatosEstudianteNotas)
}

// GetEspaciosAcademicosDocente ...
// @Title GetEspaciosAcademicosDocente
// @Description Listar la carga academica relacionada a determinado docente
// @Param	id_docente		path 	int	true		"Id docente"
// @Success 200 {}
// @Failure 404 not found resource
// @router /docentes/:id_docente/espacios-academicos [get]
func (c *NotasController) GetEspaciosAcademicosDocente() {
	defer errorhandler.HandlePanic(&c.Controller)

	idDocente := c.Ctx.Input.Param(":id_docente")

	resultados, err := services.GetEspaciosAcademicosDocente(idDocente)

	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, err.Error())
	} else {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultados)
	}

	c.ServeJSON()
}

// GetModificacionExtemporanea ...
// @Title GetModificacionExtemporanea
// @Description Chequear si hay modificacion extemporanea para la asignatura
// @Param	id_asignatura		path 	string	true		"Id asignatura"
// @Success 200 {}
// @Failure 404 not found resource
// @router /asignaturas/:id_asignatura/modificacion-extemporanea [get]
func (c *NotasController) GetModificacionExtemporanea() {
	defer errorhandler.HandlePanic(&c.Controller)
	idAsignatura := c.Ctx.Input.Param(":id_asignatura")

	registroAsignatura, err := services.GetModificacionExtemporanea(idAsignatura)

	if err == nil && fmt.Sprintf("%v", registroAsignatura["Status"]) == "200" {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 200, registroAsignatura["Data"])
	} else {
		logs.Error(err)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil)
	}

	c.ServeJSON()

}

// GetDatosDocenteAsignatura ...
// @Title GetDatosDocenteAsignatura
// @Description Obtener la informacion de docente y asingnatura solicitada
// @Param	id_asignatura		path 	string	true		"Id asignatura"
// @Success 200 {}
// @Failure 404 not found resource
// @router /asignaturas/:id_asignatura/info-docente [get]
func (c *NotasController) GetDatosDocenteAsignatura() {
	defer errorhandler.HandlePanic(&c.Controller)
	idAsignatura := c.Ctx.Input.Param(":id_asignatura")

	resultado, err := services.GetDatosDocenteAsignatura(idAsignatura)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultado)
	} else {
		logs.Error(err)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil)
	}

	c.ServeJSON()
}

// GetPorcentajesAsignatura ...
// @Title GetPorcentajesAsignatura
// @Description Obtener los porcentajes de la asignatura solicitada
// @Param	id_asignatura		path 	string	true		"Id asignatura"
// @Param	id_periodo		path 	int	true		"Id periodo"
// @Success 200 {}
// @Failure 404 not found resource
// @router /asignaturas/:id_asignatura/periodos/:id_periodo/porcentajes [get]
func (c *NotasController) GetPorcentajesAsignatura() {
	defer errorhandler.HandlePanic(&c.Controller)

	idAsignatura := c.Ctx.Input.Param(":id_asignatura")
	idPeriodo := c.Ctx.Input.Param(":id_periodo")

	resultado, err := services.GetPorcentajesAsignatura(idAsignatura, idPeriodo)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultado)
	} else {
		logs.Error(err)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil)
	}

	c.ServeJSON()
}

// PutPorcentajesAsignatura ...
// @Title PutPorcentajesAsignatura
// @Description Modificar los porcentajes de la asignatura solicitada
// @Param   body        body    {}  true        "body Modificar registro Asignatura content"
// @Success 200 {}
// @Failure 400 the request contains incorrect syntax
// @router /asignaturas/porcentajes [put]
func (c *NotasController) PutPorcentajesAsignatura() {
	defer errorhandler.HandlePanic(&c.Controller)

	data := c.Ctx.Input.RequestBody

	resultado, err := services.PutPorcentajeAsignatura(data)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultado)
	} else {
		logs.Error(err)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, err.Error())
	}

	c.ServeJSON()
}

// GetCapturaNotas ...
// @Title GetCapturaNotas
// @Description Obtener lista de estudiantes con los registros de notas para determinada asignatura
// @Param	id_asignatura	path	string	true		"Id asignatura"
// @Param	id_periodo				path	int		true		"Id periodo"
// @Success 200 {}
// @Failure 404 not found resource
// @router /notas/asignaturas/:id_asignatura/periodos/:id_periodo/estudiantes [get]
func (c *NotasController) GetCapturaNotas() {
	defer errorhandler.HandlePanic(&c.Controller)

	idEspacioAcademico := c.Ctx.Input.Param(":id_asignatura")
	idPeriodo := c.Ctx.Input.Param(":id_periodo")

	resultado, err := services.GetCapturaNotas(idEspacioAcademico, idPeriodo)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultado)
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}

	c.ServeJSON()
}

// PutCapturaNotas ...
// @Title PutCapturaNotas
// @Description Modificar registro de notas para estudiantes de determinada asignatura
// @Param   body        body    {}  true        "body Notas Estudiantes"
// @Success 200 {}
// @Failure 400 the request contains incorrect syntax
// @router /notas/asignaturas/estudiantes [put]
func (c *NotasController) PutCapturaNotas() {
	defer errorhandler.HandlePanic(&c.Controller)
	data := c.Ctx.Input.RequestBody

	resultado, err := services.PutCapturaNotas(data)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultado)
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}

	c.ServeJSON()
}

// GetEstadosRegistros ...
// @Title GetEstadosRegistros
// @Description Listar asignaturas docentes  junto estado registro
// @Param	id_periodo		path 	int	true		"Id periodo"
// @Success 200 {}
// @Failure 404 not found resource
// @router /periodos/:id_periodo/estados-registros [get]
func (c *NotasController) GetEstadosRegistros() {
	defer errorhandler.HandlePanic(&c.Controller)

	idPeriodo := c.Ctx.Input.Param(":id_periodo")

	resultado, err := services.GetEstadosRegistros(idPeriodo)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultado)
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}

	c.ServeJSON()
}

// GetDatosEstudianteNotas ...
// @Title GetDatosEstudianteNotas
// @Description Obtener la informacion de estudiante y notas asignaturas
// @Param	id_estudiante	path	int		true	"Id estudiante"
// @Success 200 {}
// @Failure 404 not found resource
// @router /notas/estudiantes/:id_estudiante [get]
func (c *NotasController) GetDatosEstudianteNotas() {
	defer errorhandler.HandlePanic(&c.Controller)

	idEstudiante := c.Ctx.Input.Param(":id_estudiante")

	resultado, err := services.GetDatosEstudianteNotas(idEstudiante)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultado)
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}

	c.ServeJSON()
}