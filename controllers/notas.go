package controllers

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	request "github.com/udistrital/sga_mid_notas/models"
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

	id_docente := c.Ctx.Input.Param(":id_docente")

	resultados := []interface{}{}

	var EspaciosAcademicosRegistros map[string]interface{}
	var proyectos []interface{}
	var calendarios []interface{}
	var periodos map[string]interface{}

	errEspaciosAcademicosRegistros := request.GetJson("http://"+beego.AppConfig.String("EspaciosAcademicosService")+"espacio-academico?query=activo:true,docente_id:"+fmt.Sprintf("%v", id_docente), &EspaciosAcademicosRegistros)
	if errEspaciosAcademicosRegistros == nil && fmt.Sprintf("%v", EspaciosAcademicosRegistros["Status"]) == "200" {

		errProyectos := request.GetJson("http://"+beego.AppConfig.String("ProyectoAcademicoService")+"proyecto_academico_institucion?query=Activo:true&fields=Id,Nombre,NivelFormacionId&limit=0", &proyectos)
		if errProyectos == nil && fmt.Sprintf("%v", proyectos[0]) != "map[]" {

			errCalendario := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario?query=Activo:true&fields=Id,Nombre,PeriodoId&limit=0", &calendarios)
			if errCalendario == nil && fmt.Sprintf("%v", calendarios[0]) != "map[]" {

				errPeriodos := request.GetJson("http://"+beego.AppConfig.String("ParametroService")+"periodo?query=Activo:true&fields=Id,Nombre&limit=0", &periodos)
				if errPeriodos == nil && fmt.Sprintf("%v", periodos["Status"]) == "200" && fmt.Sprintf("%v", periodos["Data"]) != "[map[]]" {

					for _, espacioAcademicoRegistro := range EspaciosAcademicosRegistros["Data"].([]interface{}) {

						calendario := findIdsbyId(calendarios, fmt.Sprintf("%v", espacioAcademicoRegistro.(map[string]interface{})["periodo_id"]))
						if fmt.Sprintf("%v", calendario) != "map[]" {

							proyecto := findIdsbyId(proyectos, fmt.Sprintf("%v", espacioAcademicoRegistro.(map[string]interface{})["proyecto_academico_id"]))
							if fmt.Sprintf("%v", proyecto) != "map[]" {

								resultados = append(resultados, map[string]interface{}{
									"Nivel":              proyecto["NivelFormacionId"].(map[string]interface{})["Nombre"],
									"Nivel_id":           proyecto["NivelFormacionId"].(map[string]interface{})["Id"],
									"Codigo":             espacioAcademicoRegistro.(map[string]interface{})["codigo"],
									"Asignatura":         espacioAcademicoRegistro.(map[string]interface{})["nombre"],
									"Periodo":            findNamebyId(periodos["Data"].([]interface{}), fmt.Sprintf("%v", calendario["PeriodoId"])),
									"PeriodoId":          espacioAcademicoRegistro.(map[string]interface{})["periodo_id"],
									"Grupo":              espacioAcademicoRegistro.(map[string]interface{})["grupo"],
									"Inscritos":          espacioAcademicoRegistro.(map[string]interface{})["inscritos"],
									"Proyecto_Academico": proyecto["Nombre"],
									"AsignaturaId":       espacioAcademicoRegistro.(map[string]interface{})["_id"],
								})
							}
						}
					}

					if fmt.Sprintf("%v", resultados) != "[]" {
						c.Ctx.Output.SetStatus(200)
						c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultados)
					} else {
						c.Ctx.Output.SetStatus(404)
						c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil)
					}
				} else {
					logs.Error(errPeriodos)
					c.Ctx.Output.SetStatus(404)
					c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil)
				}
			} else {
				logs.Error(errCalendario)
				c.Ctx.Output.SetStatus(404)
				c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil)
			}
		} else {
			logs.Error(errProyectos)
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil)
		}
	} else {
		logs.Error(errEspaciosAcademicosRegistros)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil)
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

	id_asignatura := c.Ctx.Input.Param(":id_asignatura")

	var RegistroAsignatura map[string]interface{}
	errRegistroAsignatura := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"registro?query=activo:true,espacio_academico_id:"+fmt.Sprintf("%v", id_asignatura)+"&fields=estado_registro_id,modificacion_extemporanea&limit=0", &RegistroAsignatura)
	if errRegistroAsignatura == nil && fmt.Sprintf("%v", RegistroAsignatura["Status"]) == "200" {

		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 200, RegistroAsignatura["Data"])
	} else {
		logs.Error(errRegistroAsignatura)
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

	id_asignatura := c.Ctx.Input.Param(":id_asignatura")

	resultado := []interface{}{}

	var EspacioAcademicoRegistro map[string]interface{}
	var DocenteInfo []map[string]interface{}
	var proyecto []interface{}
	var calendario []interface{}
	var periodo map[string]interface{}

	errEspacioAcademicoRegistro := request.GetJson("http://"+beego.AppConfig.String("EspaciosAcademicosService")+"espacio-academico/"+fmt.Sprintf("%v", id_asignatura), &EspacioAcademicoRegistro)
	if errEspacioAcademicoRegistro == nil && fmt.Sprintf("%v", EspacioAcademicoRegistro["Status"]) == "200" {

		errDocenteInfo := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"datos_identificacion?query=Activo:true,TerceroId:"+fmt.Sprintf("%v", EspacioAcademicoRegistro["Data"].(map[string]interface{})["docente_id"])+"&sortby=Id&order=desc&limit=1", &DocenteInfo)
		if errDocenteInfo == nil && fmt.Sprintf("%v", DocenteInfo[0]) != "map[]" {

			errProyecto := request.GetJson("http://"+beego.AppConfig.String("ProyectoAcademicoService")+"proyecto_academico_institucion?query=Activo:true,Id:"+fmt.Sprintf("%v", EspacioAcademicoRegistro["Data"].(map[string]interface{})["proyecto_academico_id"]), &proyecto)
			if errProyecto == nil && fmt.Sprintf("%v", proyecto[0]) != "map[]" {

				errCalendario := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario?query=Activo:true,Id:"+fmt.Sprintf("%v", EspacioAcademicoRegistro["Data"].(map[string]interface{})["periodo_id"]), &calendario)
				if errCalendario == nil && fmt.Sprintf("%v", calendario[0]) != "map[]" {

					errPeriodo := request.GetJson("http://"+beego.AppConfig.String("ParametroService")+"periodo/"+fmt.Sprintf("%v", calendario[0].(map[string]interface{})["PeriodoId"]), &periodo)
					if errPeriodo == nil && fmt.Sprintf("%v", periodo["Status"]) == "200" {

						resultado = append(resultado, map[string]interface{}{
							"Docente":        DocenteInfo[0]["TerceroId"].(map[string]interface{})["NombreCompleto"],
							"Identificacion": DocenteInfo[0]["Numero"],
							"Codigo":         EspacioAcademicoRegistro["Data"].(map[string]interface{})["codigo"],
							"Asignatura":     EspacioAcademicoRegistro["Data"].(map[string]interface{})["nombre"],
							"Nivel":          proyecto[0].(map[string]interface{})["NivelFormacionId"].(map[string]interface{})["Nombre"],
							"Grupo":          EspacioAcademicoRegistro["Data"].(map[string]interface{})["grupo"],
							"Inscritos":      EspacioAcademicoRegistro["Data"].(map[string]interface{})["inscritos"],
							"Creditos":       EspacioAcademicoRegistro["Data"].(map[string]interface{})["creditos"],
							"Periodo":        periodo["Data"].(map[string]interface{})["Nombre"],
						})
						c.Ctx.Output.SetStatus(200)
						c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultado)
					} else {
						logs.Error(errPeriodo)
						c.Ctx.Output.SetStatus(404)
						c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil)
					}
				} else {
					logs.Error(errCalendario)
					c.Ctx.Output.SetStatus(404)
					c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil)
				}
			} else {
				logs.Error(errProyecto)
				c.Ctx.Output.SetStatus(404)
				c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil)
			}
		} else {
			logs.Error(errDocenteInfo)
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil)
		}
	} else {
		logs.Error(errEspacioAcademicoRegistro)
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

	id_asignatura := c.Ctx.Input.Param(":id_asignatura")
	id_periodo := c.Ctx.Input.Param(":id_periodo")

	if InfoPorcentajes, ok := EstadosRegistroIDs(); ok {

		resultados := []interface{}{}

		var RegistroAsignatura map[string]interface{}
		errRegistroAsignatura := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"registro?query=activo:true,espacio_academico_id:"+fmt.Sprintf("%v", id_asignatura)+",periodo_id:"+fmt.Sprintf("%v", id_periodo), &RegistroAsignatura)
		if errRegistroAsignatura == nil {

			if fmt.Sprintf("%v", RegistroAsignatura["Status"]) == "200" {
				for _, PorcentajeAsignatura := range RegistroAsignatura["Data"].([]interface{}) {

					resultados = append(resultados, map[string]interface{}{
						"id":               PorcentajeAsignatura.(map[string]interface{})["_id"],
						"estadoRegistro":   PorcentajeAsignatura.(map[string]interface{})["estado_registro_id"],
						"fields":           PorcentajeAsignatura.(map[string]interface{})["estructura_nota"],
						"editExtemporaneo": PorcentajeAsignatura.(map[string]interface{})["modificacion_extemporanea"],
						"finalizado":       PorcentajeAsignatura.(map[string]interface{})["finalizado"],
					})

					IdEstado := fmt.Sprintf("%v", PorcentajeAsignatura.(map[string]interface{})["estado_registro_id"])

					if InfoPorcentajes.Corte1.IdEstado == IdEstado {
						InfoPorcentajes.Corte1.Existe = true
					}
					if InfoPorcentajes.Corte2.IdEstado == IdEstado {
						InfoPorcentajes.Corte2.Existe = true
					}
					if InfoPorcentajes.Examen.IdEstado == IdEstado {
						InfoPorcentajes.Examen.Existe = true
					}
					if InfoPorcentajes.Habilit.IdEstado == IdEstado {
						InfoPorcentajes.Habilit.Existe = true
					}
					if InfoPorcentajes.Definitiva.IdEstado == IdEstado {
						InfoPorcentajes.Definitiva.Existe = true
					}

				}
			}

			if !InfoPorcentajes.Corte1.Existe {
				resultados = append(resultados, passPorcentajeEmpty(InfoPorcentajes.Corte1.IdEstado))
			}
			if !InfoPorcentajes.Corte2.Existe {
				resultados = append(resultados, passPorcentajeEmpty(InfoPorcentajes.Corte2.IdEstado))
			}
			if !InfoPorcentajes.Examen.Existe {
				resultados = append(resultados, passPorcentajeEmpty(InfoPorcentajes.Examen.IdEstado))
			}
			if !InfoPorcentajes.Habilit.Existe {
				resultados = append(resultados, passPorcentajeEmpty(InfoPorcentajes.Habilit.IdEstado))
			}
			if !InfoPorcentajes.Definitiva.Existe {
				resultados = append(resultados, passPorcentajeEmpty(InfoPorcentajes.Definitiva.IdEstado))
			}

			c.Ctx.Output.SetStatus(200)
			c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultados)

		} else {
			logs.Error(errRegistroAsignatura)
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil)
		}
	} else {
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

	var inputData map[string]interface{}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &inputData); err == nil {

		valido := validatePutPorcentajes(inputData)

		crearRegistros := fmt.Sprintf("%v", inputData["Accion"]) == "Crear"
		guardarRegistros := fmt.Sprintf("%v", inputData["Accion"]) == "Guardar"

		if valido {

			var crearRegistrosReporte []interface{}
			var crearSalioMal = false

			var guardarRegistroReporte interface{}
			var guardarSalioMal = false

			for _, PorcentajeNota := range inputData["PorcentajesNotas"].([]interface{}) {

				id := PorcentajeNota.(map[string]interface{})["id"]
				fields := PorcentajeNota.(map[string]interface{})["fields"]
				estadoRegistro := PorcentajeNota.(map[string]interface{})["estadoRegistro"]
				editporTiempo := fmt.Sprintf("%v", PorcentajeNota.(map[string]interface{})["editporTiempo"]) == "true"
				editExtemporaneo := fmt.Sprintf("%v", PorcentajeNota.(map[string]interface{})["editExtemporaneo"]) == "true"
				finalizado := fmt.Sprintf("%v", PorcentajeNota.(map[string]interface{})["finalizado"]) == "true"

				if crearRegistros || ((estadoRegistro == inputData["Estado_registro"]) && ((!finalizado && editporTiempo) || editExtemporaneo)) {

					if fmt.Sprintf("%v", id) == "" && crearRegistros {
						formato := map[string]interface{}{
							"nombre":                    inputData["Info"].(map[string]interface{})["nombre"],
							"descripcion":               " ",
							"codigo_abreviacion":        inputData["Info"].(map[string]interface{})["codigo"],
							"codigo":                    inputData["Info"].(map[string]interface{})["codigo"],
							"periodo_id":                inputData["Info"].(map[string]interface{})["periodo"],
							"nivel_id":                  inputData["Info"].(map[string]interface{})["nivel"],
							"espacio_academico_id":      inputData["Info"].(map[string]interface{})["espacio_academico"],
							"estado_registro_id":        estadoRegistro,
							"estructura_nota":           fields,
							"finalizado":                false,
							"modificacion_extemporanea": false,
							"activo":                    true,
						}

						var PorcentajeAsignaturaNew map[string]interface{}
						errPorcentajeAsignaturaNew := request.SendJson("http://"+beego.AppConfig.String("CalificacionesService")+"registro", "POST", &PorcentajeAsignaturaNew, formato)
						if errPorcentajeAsignaturaNew == nil && fmt.Sprintf("%v", PorcentajeAsignaturaNew["Status"]) == "201" {
							crearRegistrosReporte = append(crearRegistrosReporte, PorcentajeAsignaturaNew["Data"])
						} else {
							logs.Error(errPorcentajeAsignaturaNew)
							crearSalioMal = true
						}
					} else if guardarRegistros {
						var PorcentajeAsignatura map[string]interface{}
						estructura_nota := map[string]interface{}{
							"estructura_nota": fields,
						}
						errPorcentajeAsignatura := request.SendJson("http://"+beego.AppConfig.String("CalificacionesService")+"registro/"+fmt.Sprintf("%v", id), "PUT", &PorcentajeAsignatura, estructura_nota)
						if errPorcentajeAsignatura == nil && fmt.Sprintf("%v", PorcentajeAsignatura["Status"]) == "200" {
							guardarRegistroReporte = PorcentajeAsignatura["Data"]
						} else {
							logs.Error(errPorcentajeAsignatura)
							guardarSalioMal = true
						}
					}
				}
			}

			if crearRegistros {
				if crearSalioMal {
					for _, reporte := range crearRegistrosReporte {
						id := fmt.Sprintf("%v", reporte.(map[string]interface{})["_id"])
						var PorcentajeAsignaturaDel map[string]interface{}
						errPorcentajeAsignaturaDel := request.SendJson("http://"+beego.AppConfig.String("CalificacionesService")+"registro/"+id, "DELETE", &PorcentajeAsignaturaDel, nil)
						if errPorcentajeAsignaturaDel == nil && fmt.Sprintf("%v", PorcentajeAsignaturaDel["Status"]) == "200" {
							logs.Error(PorcentajeAsignaturaDel)
						} else {
							logs.Error(errPorcentajeAsignaturaDel)
						}
					}
					c.Ctx.Output.SetStatus(400)
					c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil)
				} else {
					c.Ctx.Output.SetStatus(200)
					c.Data["json"] = requestresponse.APIResponseDTO(true, 200, crearRegistrosReporte)
				}
			} else if guardarRegistros {
				if guardarSalioMal {
					c.Ctx.Output.SetStatus(400)
					c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, "Error service PutPorcentajesAsignatura: The request contains an incorrect data type or an invalid parameter")
				} else {
					c.Ctx.Output.SetStatus(200)
					c.Data["json"] = requestresponse.APIResponseDTO(true, 200, guardarRegistroReporte)
				}
			} else {
				c.Ctx.Output.SetStatus(400)
				c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, "Error service PutPorcentajesAsignatura: The request contains an incorrect data type or an invalid parameter")
			}
		} else {
			c.Ctx.Output.SetStatus(400)
			c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, "Error service PutPorcentajesAsignatura: The request contains an incorrect data type or an invalid parameter")
		}
	} else {
		logs.Error(err)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, "Error service PutPorcentajesAsignatura: The request contains an incorrect data type or an invalid parameter")
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

	id_espacio_academico := c.Ctx.Input.Param(":id_asignatura")
	id_periodo := c.Ctx.Input.Param(":id_periodo")

	var resultado map[string]interface{}
	datos := []interface{}{}

	var EspaciosAcademicosEstudiantes map[string]interface{}
	var RegistroCalificacion map[string]interface{}
	var EstudianteInformacion []interface{}

	if InformacionCalificaciones, ok := EstadosRegistroIDs(); ok {

		errRegistroCalificacion := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"registro?query=activo:true,periodo_id:"+fmt.Sprintf("%v", id_periodo)+",espacio_academico_id:"+fmt.Sprintf("%v", id_espacio_academico), &RegistroCalificacion)
		if errRegistroCalificacion == nil && fmt.Sprintf("%v", RegistroCalificacion["Status"]) == "200" {
			for _, EstadosRegistro := range RegistroCalificacion["Data"].([]interface{}) {
				if fmt.Sprintf("%v", EstadosRegistro.(map[string]interface{})["estado_registro_id"]) == InformacionCalificaciones.Corte1.IdEstado {
					InformacionCalificaciones.Corte1.Existe = true
					InformacionCalificaciones.Corte1.IdRegistroNota = fmt.Sprintf("%v", EstadosRegistro.(map[string]interface{})["_id"])
					InformacionCalificaciones.Corte1.Finalizado = fmt.Sprintf("%v", EstadosRegistro.(map[string]interface{})["finalizado"]) == "true"
				}
				if fmt.Sprintf("%v", EstadosRegistro.(map[string]interface{})["estado_registro_id"]) == InformacionCalificaciones.Corte2.IdEstado {
					InformacionCalificaciones.Corte2.Existe = true
					InformacionCalificaciones.Corte2.IdRegistroNota = fmt.Sprintf("%v", EstadosRegistro.(map[string]interface{})["_id"])
					InformacionCalificaciones.Corte2.Finalizado = fmt.Sprintf("%v", EstadosRegistro.(map[string]interface{})["finalizado"]) == "true"
				}
				if fmt.Sprintf("%v", EstadosRegistro.(map[string]interface{})["estado_registro_id"]) == InformacionCalificaciones.Examen.IdEstado {
					InformacionCalificaciones.Examen.Existe = true
					InformacionCalificaciones.Examen.IdRegistroNota = fmt.Sprintf("%v", EstadosRegistro.(map[string]interface{})["_id"])
					InformacionCalificaciones.Examen.Finalizado = fmt.Sprintf("%v", EstadosRegistro.(map[string]interface{})["finalizado"]) == "true"
				}
				if fmt.Sprintf("%v", EstadosRegistro.(map[string]interface{})["estado_registro_id"]) == InformacionCalificaciones.Habilit.IdEstado {
					InformacionCalificaciones.Habilit.Existe = true
					InformacionCalificaciones.Habilit.IdRegistroNota = fmt.Sprintf("%v", EstadosRegistro.(map[string]interface{})["_id"])
					InformacionCalificaciones.Habilit.Finalizado = fmt.Sprintf("%v", EstadosRegistro.(map[string]interface{})["finalizado"]) == "true"
				}
				if fmt.Sprintf("%v", EstadosRegistro.(map[string]interface{})["estado_registro_id"]) == InformacionCalificaciones.Definitiva.IdEstado {
					InformacionCalificaciones.Definitiva.Existe = true
					InformacionCalificaciones.Definitiva.IdRegistroNota = fmt.Sprintf("%v", EstadosRegistro.(map[string]interface{})["_id"])
					InformacionCalificaciones.Definitiva.Finalizado = fmt.Sprintf("%v", EstadosRegistro.(map[string]interface{})["finalizado"]) == "true"
				}
			}

			errEspaciosAcademicosEstudiantes := request.GetJson("http://"+beego.AppConfig.String("EspaciosAcademicosService")+"espacio-academico-estudiantes?query=activo:true,espacio_academico_id:"+fmt.Sprintf("%v", id_espacio_academico)+",periodo_id:"+fmt.Sprintf("%v", id_periodo), &EspaciosAcademicosEstudiantes)
			if errEspaciosAcademicosEstudiantes == nil && fmt.Sprintf("%v", EspaciosAcademicosEstudiantes["Status"]) == "200" {

				for _, espaciosAcademicoEstudiante := range EspaciosAcademicosEstudiantes["Data"].([]interface{}) {

					id_estudiante := espaciosAcademicoEstudiante.(map[string]interface{})["estudiante_id"]

					errEstudianteInformacion := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"info_complementaria_tercero?query=InfoComplementariaId.Id:93,TerceroId.Id:"+fmt.Sprintf("%v", id_estudiante), &EstudianteInformacion)
					if errEstudianteInformacion == nil && fmt.Sprintf("%v", EstudianteInformacion[0]) != "map[]" {

						Codigo := EstudianteInformacion[0].(map[string]interface{})["Dato"]
						Nombre1 := EstudianteInformacion[0].(map[string]interface{})["TerceroId"].(map[string]interface{})["PrimerNombre"]
						Nombre2 := EstudianteInformacion[0].(map[string]interface{})["TerceroId"].(map[string]interface{})["SegundoNombre"]
						Apellido1 := EstudianteInformacion[0].(map[string]interface{})["TerceroId"].(map[string]interface{})["PrimerApellido"]
						Apellido2 := EstudianteInformacion[0].(map[string]interface{})["TerceroId"].(map[string]interface{})["SegundoApellido"]

						if InformacionCalificaciones.Corte1.Existe {
							var InfoNota map[string]interface{}
							errInfoNota := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota?query=activo:true,registro_id:"+InformacionCalificaciones.Corte1.IdRegistroNota+",estudiante_id:"+fmt.Sprintf("%v", id_estudiante), &InfoNota)
							if errInfoNota == nil && fmt.Sprintf("%v", InfoNota["Status"]) == "200" {
								InformacionCalificaciones.Corte1.informacion = passNotaInf(InfoNota)
							} else {
								InformacionCalificaciones.Corte1.informacion = passNotaEmpty()
							}
						} else {
							InformacionCalificaciones.Corte1.informacion = passNotaEmpty()
						}

						if InformacionCalificaciones.Corte2.Existe {
							var InfoNota map[string]interface{}
							errInfoNota := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota?query=activo:true,registro_id:"+InformacionCalificaciones.Corte2.IdRegistroNota+",estudiante_id:"+fmt.Sprintf("%v", id_estudiante), &InfoNota)
							if errInfoNota == nil && fmt.Sprintf("%v", InfoNota["Status"]) == "200" {
								InformacionCalificaciones.Corte2.informacion = passNotaInf(InfoNota)
							} else {
								InformacionCalificaciones.Corte2.informacion = passNotaEmpty()
							}
						} else {
							InformacionCalificaciones.Corte2.informacion = passNotaEmpty()
						}

						if InformacionCalificaciones.Examen.Existe {
							var InfoNota map[string]interface{}
							errInfoNota := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota?query=activo:true,registro_id:"+InformacionCalificaciones.Examen.IdRegistroNota+",estudiante_id:"+fmt.Sprintf("%v", id_estudiante), &InfoNota)
							if errInfoNota == nil && fmt.Sprintf("%v", InfoNota["Status"]) == "200" {
								InformacionCalificaciones.Examen.informacion = passNotaInf(InfoNota)
							} else {
								InformacionCalificaciones.Examen.informacion = passNotaEmpty()
							}
						} else {
							InformacionCalificaciones.Examen.informacion = passNotaEmpty()
						}

						if InformacionCalificaciones.Habilit.Existe {
							var InfoNota map[string]interface{}
							errInfoNota := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota?query=activo:true,registro_id:"+InformacionCalificaciones.Habilit.IdRegistroNota+",estudiante_id:"+fmt.Sprintf("%v", id_estudiante), &InfoNota)
							if errInfoNota == nil && fmt.Sprintf("%v", InfoNota["Status"]) == "200" {
								InformacionCalificaciones.Habilit.informacion = passNotaInf(InfoNota)
							} else {
								InformacionCalificaciones.Habilit.informacion = passNotaEmpty()
							}
						} else {
							InformacionCalificaciones.Habilit.informacion = passNotaEmpty()
						}

						if InformacionCalificaciones.Definitiva.Existe {
							var InfoNota map[string]interface{}
							errInfoNota := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota?query=activo:true,registro_id:"+InformacionCalificaciones.Definitiva.IdRegistroNota+",estudiante_id:"+fmt.Sprintf("%v", id_estudiante), &InfoNota)
							if errInfoNota == nil && fmt.Sprintf("%v", InfoNota["Status"]) == "200" {
								InformacionCalificaciones.Definitiva.informacion = passNotaInf(InfoNota)
							} else {
								InformacionCalificaciones.Definitiva.informacion = passNotaEmpty()
							}
						} else {
							InformacionCalificaciones.Definitiva.informacion = passNotaEmpty()
						}

						datos = append(datos, map[string]interface{}{
							"Id":         id_estudiante,
							"Codigo":     Codigo,
							"Nombre":     fmt.Sprintf("%v", Nombre1) + " " + fmt.Sprintf("%v", Nombre2),
							"Apellido":   fmt.Sprintf("%v", Apellido1) + " " + fmt.Sprintf("%v", Apellido2),
							"Corte1":     InformacionCalificaciones.Corte1.informacion,
							"Corte2":     InformacionCalificaciones.Corte2.informacion,
							"Examen":     InformacionCalificaciones.Examen.informacion,
							"Habilit":    InformacionCalificaciones.Habilit.informacion,
							"Definitiva": InformacionCalificaciones.Definitiva.informacion,
							"Acumulado":  calculoAcumuladoNotas(InformacionCalificaciones),
						})

					}

				}

				var estado_registro_editable string
				if InformacionCalificaciones.Habilit.Finalizado {
					estado_registro_editable = InformacionCalificaciones.Definitiva.IdEstado
				} else if InformacionCalificaciones.Examen.Finalizado {
					estado_registro_editable = InformacionCalificaciones.Habilit.IdEstado
				} else if InformacionCalificaciones.Corte2.Finalizado {
					estado_registro_editable = InformacionCalificaciones.Examen.IdEstado
				} else if InformacionCalificaciones.Corte1.Finalizado {
					estado_registro_editable = InformacionCalificaciones.Corte2.IdEstado
				} else {
					estado_registro_editable = InformacionCalificaciones.Corte1.IdEstado
				}

				resultado = map[string]interface{}{
					"estado_registro_editable":   estado_registro_editable,
					"calificaciones_estudiantes": datos,
				}

				c.Ctx.Output.SetStatus(200)
				c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultado)

			} else {
				logs.Error(errEspaciosAcademicosEstudiantes)
				c.Ctx.Output.SetStatus(404)
				c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, "Error service GetCapturaNotas: The request contains an incorrect parameter or no record exist")
			}
		} else {
			logs.Error(errRegistroCalificacion)
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, "Error service GetCapturaNotas: The request contains an incorrect parameter or no record exist")
		}
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, "Error service GetCapturaNotas: The request contains an incorrect parameter or no record exist")
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

	var inputData map[string]interface{}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &inputData); err == nil {

		valido := validatePutNotasEstudiantes(inputData)

		if valido {

			if InfoCalificaciones, ok := EstadosRegistroIDs(); ok {

				espacioAcademico := inputData["Espacio_academico"]
				periodo := inputData["Periodo"]

				var InfoRegistro map[string]interface{}
				errInfoRegistro := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"registro?query=activo:true,periodo_id:"+fmt.Sprintf("%v", periodo)+",espacio_academico_id:"+fmt.Sprintf("%v", espacioAcademico), &InfoRegistro)
				if errInfoRegistro == nil && fmt.Sprintf("%v", InfoRegistro["Status"]) == "200" {

					for _, registro := range InfoRegistro["Data"].([]interface{}) {

						IdEstado := fmt.Sprintf("%v", registro.(map[string]interface{})["estado_registro_id"])
						reg := fmt.Sprintf("%v", registro.(map[string]interface{})["_id"])
						finalizado := fmt.Sprintf("%v", registro.(map[string]interface{})["finalizado"]) == "true"
						extemporaneo := fmt.Sprintf("%v", registro.(map[string]interface{})["modificacion_extemporanea"]) == "true"

						if InfoCalificaciones.Corte1.IdEstado == IdEstado {
							InfoCalificaciones.Corte1.IdRegistroNota = reg
							InfoCalificaciones.Corte1.Finalizado = finalizado
							InfoCalificaciones.Corte1.EditExtemporaneo = extemporaneo
						}
						if InfoCalificaciones.Corte2.IdEstado == IdEstado {
							InfoCalificaciones.Corte2.IdRegistroNota = reg
							InfoCalificaciones.Corte2.Finalizado = finalizado
							InfoCalificaciones.Corte2.EditExtemporaneo = extemporaneo
						}
						if InfoCalificaciones.Examen.IdEstado == IdEstado {
							InfoCalificaciones.Examen.IdRegistroNota = reg
							InfoCalificaciones.Examen.Finalizado = finalizado
							InfoCalificaciones.Examen.EditExtemporaneo = extemporaneo
						}
						if InfoCalificaciones.Habilit.IdEstado == IdEstado {
							InfoCalificaciones.Habilit.IdRegistroNota = reg
							InfoCalificaciones.Habilit.Finalizado = finalizado
							InfoCalificaciones.Habilit.EditExtemporaneo = extemporaneo
						}
						if InfoCalificaciones.Definitiva.IdEstado == IdEstado {
							InfoCalificaciones.Definitiva.IdRegistroNota = reg
							InfoCalificaciones.Definitiva.Finalizado = finalizado
							InfoCalificaciones.Definitiva.EditExtemporaneo = extemporaneo
						}
					}

					crearRegistros := fmt.Sprintf("%v", inputData["Accion"]) == "Crear"
					guardarRegistros := fmt.Sprintf("%v", inputData["Accion"]) == "Guardar" || fmt.Sprintf("%v", inputData["Accion"]) == "Cerrar"

					var crearNotasReporte []interface{}
					var crearSalioMal = false

					var respaldoNotas []interface{}
					var guardarNotasReporte []interface{}
					var guardarSalioMal = false

					for _, CalificacionEstudiante := range inputData["CalificacionesEstudiantes"].([]interface{}) {

						if crearRegistros || ((fmt.Sprintf("%v", inputData["Estado_registro"]) == InfoCalificaciones.Corte1.IdEstado) && (!InfoCalificaciones.Corte1.Finalizado || InfoCalificaciones.Corte1.EditExtemporaneo)) {

							idnota := CalificacionEstudiante.(map[string]interface{})["Corte1"].(map[string]interface{})["id"]
							nota_json := CalificacionEstudiante.(map[string]interface{})["Corte1"].(map[string]interface{})["data"]
							fallas := CalificacionEstudiante.(map[string]interface{})["Varios"].(map[string]interface{})["Fallas"]
							Observ := CalificacionEstudiante.(map[string]interface{})["Varios"].(map[string]interface{})["Observacion"]

							if fmt.Sprintf("%v", idnota) == "" {
								nota_json.(map[string]interface{})["nombre"] = inputData["Nombre"]
								nota_json.(map[string]interface{})["descripcion"] = " "
								nota_json.(map[string]interface{})["estudiante_id"] = CalificacionEstudiante.(map[string]interface{})["Id"]
								nota_json.(map[string]interface{})["registro_id"] = InfoCalificaciones.Corte1.IdRegistroNota
								nota_json.(map[string]interface{})["aprobado"] = false
								nota_json.(map[string]interface{})["homologado"] = false
								nota_json.(map[string]interface{})["activo"] = true
								var NotaNew map[string]interface{}
								errNotaNew := request.SendJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota", "POST", &NotaNew, nota_json)
								if errNotaNew == nil && fmt.Sprintf("%v", NotaNew["Status"]) == "201" {
									crearNotasReporte = append(crearNotasReporte, NotaNew["Data"])
								} else {
									logs.Error(errNotaNew)
									crearSalioMal = true
								}
							} else if guardarRegistros {

								nota_json = calculoNotasPorCortes(nota_json.(map[string]interface{}))
								nota_json.(map[string]interface{})["fallas"] = fallas
								nota_json.(map[string]interface{})["observacion_nota_id"] = Observ

								var respaldo map[string]interface{}
								errrespaldo := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota/"+fmt.Sprintf("%v", idnota), &respaldo)
								if errrespaldo == nil && fmt.Sprintf("%v", respaldo["Status"]) == "200" {
									respaldoNotas = append(respaldoNotas, respaldo["Data"])
								} else {
									guardarSalioMal = true
								}

								var Nota map[string]interface{}
								errNota := request.SendJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota/"+fmt.Sprintf("%v", idnota), "PUT", &Nota, nota_json)
								if errNota == nil && fmt.Sprintf("%v", Nota["Status"]) == "200" {
									guardarNotasReporte = append(guardarNotasReporte, Nota["Data"])
								} else {
									logs.Error(errNota)
									guardarSalioMal = true
								}
							}
						}

						if crearRegistros || ((fmt.Sprintf("%v", inputData["Estado_registro"]) == InfoCalificaciones.Corte2.IdEstado) && (!InfoCalificaciones.Corte2.Finalizado || InfoCalificaciones.Corte2.EditExtemporaneo)) {

							idnota := CalificacionEstudiante.(map[string]interface{})["Corte2"].(map[string]interface{})["id"]
							nota_json := CalificacionEstudiante.(map[string]interface{})["Corte2"].(map[string]interface{})["data"]
							fallas := CalificacionEstudiante.(map[string]interface{})["Varios"].(map[string]interface{})["Fallas"]
							Observ := CalificacionEstudiante.(map[string]interface{})["Varios"].(map[string]interface{})["Observacion"]

							if fmt.Sprintf("%v", idnota) == "" {
								nota_json.(map[string]interface{})["nombre"] = inputData["Nombre"]
								nota_json.(map[string]interface{})["descripcion"] = " "
								nota_json.(map[string]interface{})["estudiante_id"] = CalificacionEstudiante.(map[string]interface{})["Id"]
								nota_json.(map[string]interface{})["registro_id"] = InfoCalificaciones.Corte2.IdRegistroNota
								nota_json.(map[string]interface{})["aprobado"] = false
								nota_json.(map[string]interface{})["homologado"] = false
								nota_json.(map[string]interface{})["activo"] = true
								var NotaNew map[string]interface{}
								errNotaNew := request.SendJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota", "POST", &NotaNew, nota_json)
								if errNotaNew == nil && fmt.Sprintf("%v", NotaNew["Status"]) == "201" {
									crearNotasReporte = append(crearNotasReporte, NotaNew["Data"])
								} else {
									logs.Error(errNotaNew)
									crearSalioMal = true
								}
							} else if guardarRegistros {

								nota_json = calculoNotasPorCortes(nota_json.(map[string]interface{}))
								nota_json.(map[string]interface{})["fallas"] = fallas
								nota_json.(map[string]interface{})["observacion_nota_id"] = Observ

								var respaldo map[string]interface{}
								errrespaldo := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota/"+fmt.Sprintf("%v", idnota), &respaldo)
								if errrespaldo == nil && fmt.Sprintf("%v", respaldo["Status"]) == "200" {
									respaldoNotas = append(respaldoNotas, respaldo["Data"])
								} else {
									guardarSalioMal = true
								}

								var Nota map[string]interface{}
								errNota := request.SendJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota/"+fmt.Sprintf("%v", idnota), "PUT", &Nota, nota_json)
								if errNota == nil && fmt.Sprintf("%v", Nota["Status"]) == "200" {
									guardarNotasReporte = append(guardarNotasReporte, Nota["Data"])
								} else {
									logs.Error(errNota)
									guardarSalioMal = true
								}
							}
						}

						if crearRegistros || ((fmt.Sprintf("%v", inputData["Estado_registro"]) == InfoCalificaciones.Examen.IdEstado) && (!InfoCalificaciones.Examen.Finalizado || InfoCalificaciones.Examen.EditExtemporaneo)) {

							idnota := CalificacionEstudiante.(map[string]interface{})["Examen"].(map[string]interface{})["id"]
							nota_json := CalificacionEstudiante.(map[string]interface{})["Examen"].(map[string]interface{})["data"]
							fallas := CalificacionEstudiante.(map[string]interface{})["Varios"].(map[string]interface{})["Fallas"]
							Observ := CalificacionEstudiante.(map[string]interface{})["Varios"].(map[string]interface{})["Observacion"]

							if fmt.Sprintf("%v", idnota) == "" {
								nota_json.(map[string]interface{})["nombre"] = inputData["Nombre"]
								nota_json.(map[string]interface{})["descripcion"] = " "
								nota_json.(map[string]interface{})["estudiante_id"] = CalificacionEstudiante.(map[string]interface{})["Id"]
								nota_json.(map[string]interface{})["registro_id"] = InfoCalificaciones.Examen.IdRegistroNota
								nota_json.(map[string]interface{})["aprobado"] = false
								nota_json.(map[string]interface{})["homologado"] = false
								nota_json.(map[string]interface{})["activo"] = true
								var NotaNew map[string]interface{}
								errNotaNew := request.SendJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota", "POST", &NotaNew, nota_json)
								if errNotaNew == nil && fmt.Sprintf("%v", NotaNew["Status"]) == "201" {
									crearNotasReporte = append(crearNotasReporte, NotaNew["Data"])
								} else {
									logs.Error(errNotaNew)
									crearSalioMal = true
								}
							} else if guardarRegistros {

								nota_json = calculoNotasPorCortes(nota_json.(map[string]interface{}))
								nota_json.(map[string]interface{})["fallas"] = fallas
								nota_json.(map[string]interface{})["observacion_nota_id"] = Observ

								var respaldo map[string]interface{}
								errrespaldo := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota/"+fmt.Sprintf("%v", idnota), &respaldo)
								if errrespaldo == nil && fmt.Sprintf("%v", respaldo["Status"]) == "200" {
									respaldoNotas = append(respaldoNotas, respaldo["Data"])
								} else {
									guardarSalioMal = true
								}

								var Nota map[string]interface{}
								errNota := request.SendJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota/"+fmt.Sprintf("%v", idnota), "PUT", &Nota, nota_json)
								if errNota == nil && fmt.Sprintf("%v", Nota["Status"]) == "200" {
									guardarNotasReporte = append(guardarNotasReporte, Nota["Data"])
								} else {
									logs.Error(errNota)
									guardarSalioMal = true
								}
							}
						}

						if crearRegistros || ((fmt.Sprintf("%v", inputData["Estado_registro"]) == InfoCalificaciones.Habilit.IdEstado) && (!InfoCalificaciones.Habilit.Finalizado || InfoCalificaciones.Habilit.EditExtemporaneo)) {

							idnota := CalificacionEstudiante.(map[string]interface{})["Habilit"].(map[string]interface{})["id"]
							nota_json := CalificacionEstudiante.(map[string]interface{})["Habilit"].(map[string]interface{})["data"]
							fallas := CalificacionEstudiante.(map[string]interface{})["Varios"].(map[string]interface{})["Fallas"]
							Observ := CalificacionEstudiante.(map[string]interface{})["Varios"].(map[string]interface{})["Observacion"]

							if fmt.Sprintf("%v", idnota) == "" {
								nota_json.(map[string]interface{})["nombre"] = inputData["Nombre"]
								nota_json.(map[string]interface{})["descripcion"] = " "
								nota_json.(map[string]interface{})["estudiante_id"] = CalificacionEstudiante.(map[string]interface{})["Id"]
								nota_json.(map[string]interface{})["registro_id"] = InfoCalificaciones.Habilit.IdRegistroNota
								nota_json.(map[string]interface{})["aprobado"] = false
								nota_json.(map[string]interface{})["homologado"] = false
								nota_json.(map[string]interface{})["activo"] = true
								var NotaNew map[string]interface{}
								errNotaNew := request.SendJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota", "POST", &NotaNew, nota_json)
								if errNotaNew == nil && fmt.Sprintf("%v", NotaNew["Status"]) == "201" {
									crearNotasReporte = append(crearNotasReporte, NotaNew["Data"])
								} else {
									logs.Error(errNotaNew)
									crearSalioMal = true
								}
							} else if guardarRegistros {

								nota_json = calculoNotasPorCortes(nota_json.(map[string]interface{}))
								nota_json.(map[string]interface{})["fallas"] = fallas
								nota_json.(map[string]interface{})["observacion_nota_id"] = Observ

								var respaldo map[string]interface{}
								errrespaldo := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota/"+fmt.Sprintf("%v", idnota), &respaldo)
								if errrespaldo == nil && fmt.Sprintf("%v", respaldo["Status"]) == "200" {
									respaldoNotas = append(respaldoNotas, respaldo["Data"])
								} else {
									guardarSalioMal = true
								}

								var Nota map[string]interface{}
								errNota := request.SendJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota/"+fmt.Sprintf("%v", idnota), "PUT", &Nota, nota_json)
								if errNota == nil && fmt.Sprintf("%v", Nota["Status"]) == "200" {
									guardarNotasReporte = append(guardarNotasReporte, Nota["Data"])
								} else {
									logs.Error(errNota)
									guardarSalioMal = true
								}
							}
						}

						if crearRegistros || ((fmt.Sprintf("%v", inputData["Estado_registro"]) == InfoCalificaciones.Definitiva.IdEstado) && (!InfoCalificaciones.Definitiva.Finalizado || InfoCalificaciones.Definitiva.EditExtemporaneo)) {

							idnota := CalificacionEstudiante.(map[string]interface{})["Definitiva"].(map[string]interface{})["id"]
							nota_json := CalificacionEstudiante.(map[string]interface{})["Definitiva"].(map[string]interface{})["data"]
							fallas := CalificacionEstudiante.(map[string]interface{})["Varios"].(map[string]interface{})["Fallas"]
							Observ := CalificacionEstudiante.(map[string]interface{})["Varios"].(map[string]interface{})["Observacion"]
							ObservCod := CalificacionEstudiante.(map[string]interface{})["Varios"].(map[string]interface{})["ObservacionCod"].(map[string]interface{})["CodigoAbreviacion"]

							if fmt.Sprintf("%v", idnota) == "" {
								nota_json.(map[string]interface{})["nombre"] = inputData["Nombre"]
								nota_json.(map[string]interface{})["descripcion"] = " "
								nota_json.(map[string]interface{})["estudiante_id"] = CalificacionEstudiante.(map[string]interface{})["Id"]
								nota_json.(map[string]interface{})["registro_id"] = InfoCalificaciones.Definitiva.IdRegistroNota
								nota_json.(map[string]interface{})["aprobado"] = false
								nota_json.(map[string]interface{})["homologado"] = false
								nota_json.(map[string]interface{})["activo"] = true
								var NotaNew map[string]interface{}
								errNotaNew := request.SendJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota", "POST", &NotaNew, nota_json)
								if errNotaNew == nil && fmt.Sprintf("%v", NotaNew["Status"]) == "201" {
									crearNotasReporte = append(crearNotasReporte, NotaNew["Data"])
								} else {
									logs.Error(errNotaNew)
									crearSalioMal = true
								}
							} else if guardarRegistros {

								def := 0.0
								nota_json = calculoNotasPorCortes(nota_json.(map[string]interface{}))
								nota_json.(map[string]interface{})["fallas"] = fallas
								nota_json.(map[string]interface{})["observacion_nota_id"] = Observ
								if fmt.Sprintf("%v", ObservCod) == "3" {
									nota_json.(map[string]interface{})["nota_definitiva"] = 0
									nota_json.(map[string]interface{})["aprobado"] = false
								} else {
									def = calculoDefinitiva(CalificacionEstudiante)
									nota_json.(map[string]interface{})["nota_definitiva"] = def
									if def >= 3 {
										nota_json.(map[string]interface{})["aprobado"] = true
									} else {
										nota_json.(map[string]interface{})["aprobado"] = false
									}
								}
								nota_json.(map[string]interface{})["valor_nota"].([]interface{})[0].(map[string]interface{})["value"] = def

								var respaldo map[string]interface{}
								errrespaldo := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota/"+fmt.Sprintf("%v", idnota), &respaldo)
								if errrespaldo == nil && fmt.Sprintf("%v", respaldo["Status"]) == "200" {
									respaldoNotas = append(respaldoNotas, respaldo["Data"])
								} else {
									guardarSalioMal = true
								}

								var Nota map[string]interface{}
								errNota := request.SendJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota/"+fmt.Sprintf("%v", idnota), "PUT", &Nota, nota_json)
								if errNota == nil && fmt.Sprintf("%v", Nota["Status"]) == "200" {
									guardarNotasReporte = append(guardarNotasReporte, Nota["Data"])
								} else {
									logs.Error(errNota)
									guardarSalioMal = true
								}
							}
						}
					}

					if crearRegistros {
						if crearSalioMal {
							for _, reporte := range crearNotasReporte {
								id := fmt.Sprintf("%v", reporte.(map[string]interface{})["_id"])
								var NotaEstudianteDel map[string]interface{}
								errNotaEstudianteDel := request.SendJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota/"+id, "DELETE", &NotaEstudianteDel, nil)
								if errNotaEstudianteDel == nil && fmt.Sprintf("%v", NotaEstudianteDel["Status"]) == "200" {
									//logs.Error(errNotaEstudianteDel)
								} else {
									logs.Error(errNotaEstudianteDel)
								}
							}
							c.Ctx.Output.SetStatus(400)
							c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, "Error service PutCapturaNotas: The request contains an incorrect data type or an invalid parameter")
						} else {
							c.Ctx.Output.SetStatus(200)
							c.Data["json"] = requestresponse.APIResponseDTO(true, 200, crearNotasReporte)
						}
					} else if guardarRegistros {

						if !guardarSalioMal {

							var cambiarRegistro struct {
								modificar bool
								campos    map[string]interface{}
							}
							cambiarRegistro.modificar = false

							if fmt.Sprintf("%v", inputData["Accion"]) == "Guardar" {
								if InfoCalificaciones.Corte1.EditExtemporaneo || InfoCalificaciones.Corte2.EditExtemporaneo || InfoCalificaciones.Examen.EditExtemporaneo || InfoCalificaciones.Habilit.EditExtemporaneo || InfoCalificaciones.Definitiva.EditExtemporaneo {
									cambiarRegistro.campos = map[string]interface{}{"finalizado": true, "modificacion_extemporanea": false}
									cambiarRegistro.modificar = true
								}
							}

							if fmt.Sprintf("%v", inputData["Accion"]) == "Cerrar" {
								cambiarRegistro.campos = map[string]interface{}{"finalizado": true, "modificacion_extemporanea": false}
								cambiarRegistro.modificar = true
							}

							if cambiarRegistro.modificar {

								idReg := ""
								if InfoCalificaciones.Corte1.IdEstado == fmt.Sprintf("%v", inputData["Estado_registro"]) {
									idReg = InfoCalificaciones.Corte1.IdRegistroNota
								}
								if InfoCalificaciones.Corte2.IdEstado == fmt.Sprintf("%v", inputData["Estado_registro"]) {
									idReg = InfoCalificaciones.Corte2.IdRegistroNota
								}
								if InfoCalificaciones.Examen.IdEstado == fmt.Sprintf("%v", inputData["Estado_registro"]) {
									idReg = InfoCalificaciones.Examen.IdRegistroNota
								}
								if InfoCalificaciones.Habilit.IdEstado == fmt.Sprintf("%v", inputData["Estado_registro"]) {
									idReg = InfoCalificaciones.Habilit.IdRegistroNota
								}
								if InfoCalificaciones.Definitiva.IdEstado == fmt.Sprintf("%v", inputData["Estado_registro"]) {
									idReg = InfoCalificaciones.Definitiva.IdRegistroNota
								}

								var RegistroCerrado map[string]interface{}
								errRegistroCerrado := request.SendJson("http://"+beego.AppConfig.String("CalificacionesService")+"registro/"+fmt.Sprintf("%v", idReg), "PUT", &RegistroCerrado, cambiarRegistro.campos)
								if errRegistroCerrado == nil && fmt.Sprintf("%v", RegistroCerrado["Status"]) == "200" {
									guardarNotasReporte = append(guardarNotasReporte, RegistroCerrado["Data"])
								} else {
									logs.Error(errRegistroCerrado)
									guardarSalioMal = true
								}
							}

						}

						if guardarSalioMal {
							for _, respaldo := range respaldoNotas {
								id := fmt.Sprintf("%v", respaldo.(map[string]interface{})["_id"])
								respaldo.(map[string]interface{})["registro_id"] = respaldo.(map[string]interface{})["registro_id"].(map[string]interface{})["_id"]
								var NotaEstudianteRespaldo map[string]interface{}
								errNotaEstudianteRespaldo := request.SendJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota/"+id, "PUT", &NotaEstudianteRespaldo, respaldo)
								if errNotaEstudianteRespaldo == nil && fmt.Sprintf("%v", NotaEstudianteRespaldo["Status"]) == "200" {
									//logs.Error(errNotaEstudianteRespaldo)
								} else {
									logs.Error(errNotaEstudianteRespaldo)
								}
							}
							c.Ctx.Output.SetStatus(400)
							c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, "Error service PutCapturaNotas: The request contains an incorrect data type or an invalid parameter")
						} else {
							c.Ctx.Output.SetStatus(200)
							c.Data["json"] = requestresponse.APIResponseDTO(true, 200, guardarNotasReporte)
						}
					}
				} else {
					logs.Error(errInfoRegistro)
					c.Ctx.Output.SetStatus(400)
					c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, "Error service PutCapturaNotas: The request contains an incorrect data type or an invalid parameter")
				}
			} else {
				c.Ctx.Output.SetStatus(400)
				c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, "Error service PutCapturaNotas: The request contains an incorrect data type or an invalid parameter")
			}
		} else {
			c.Ctx.Output.SetStatus(400)
			c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, "Error service PutCapturaNotas: The request contains an incorrect data type or an invalid parameter")
		}
	} else {
		logs.Error(err)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, "Error service PutCapturaNotas: The request contains an incorrect data type or an invalid parameter")
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

	id_periodo := c.Ctx.Input.Param(":id_periodo")

	resultados := []interface{}{}

	var EspaciosAcademicosRegistros map[string]interface{}
	var proyectos []interface{}

	var notOk bool = false

	if InfoCalificaciones, ok := EstadosRegistroIDs(); ok {
		errEspaciosAcademicosRegistros := request.GetJson("http://"+beego.AppConfig.String("EspaciosAcademicosService")+"espacio-academico?query=activo:true,periodo_id:"+fmt.Sprintf("%v", id_periodo)+"&limit=0", &EspaciosAcademicosRegistros)
		if errEspaciosAcademicosRegistros == nil && fmt.Sprintf("%v", EspaciosAcademicosRegistros["Status"]) == "200" {

			errProyectos := request.GetJson("http://"+beego.AppConfig.String("ProyectoAcademicoService")+"proyecto_academico_institucion?query=Activo:true&fields=Id,Nombre,NivelFormacionId&limit=0", &proyectos)
			if errProyectos == nil && fmt.Sprintf("%v", proyectos[0]) != "map[]" {

				for _, asignatura := range EspaciosAcademicosRegistros["Data"].([]interface{}) {

					docente_id := asignatura.(map[string]interface{})["docente_id"]
					espacioAcademico := asignatura.(map[string]interface{})["_id"]

					var DocenteInfo []map[string]interface{}
					errDocenteInfo := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"datos_identificacion?query=Activo:true,TerceroId:"+fmt.Sprintf("%v", docente_id)+"&sortby=Id&order=desc&limit=1", &DocenteInfo)
					if errDocenteInfo == nil && fmt.Sprintf("%v", DocenteInfo[0]) != "map[]" {

						var InfoRegistro map[string]interface{}
						errInfoRegistro := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"registro?query=activo:true,periodo_id:"+fmt.Sprintf("%v", id_periodo)+",espacio_academico_id:"+fmt.Sprintf("%v", espacioAcademico), &InfoRegistro)
						if errInfoRegistro == nil && fmt.Sprintf("%v", InfoRegistro["Status"]) == "200" {

							for _, registro := range InfoRegistro["Data"].([]interface{}) {
								IdEstado := fmt.Sprintf("%v", registro.(map[string]interface{})["estado_registro_id"])
								finalizado := fmt.Sprintf("%v", registro.(map[string]interface{})["finalizado"]) == "true"
								if InfoCalificaciones.Corte1.IdEstado == IdEstado {
									InfoCalificaciones.Corte1.Finalizado = finalizado
								}
								if InfoCalificaciones.Corte2.IdEstado == IdEstado {
									InfoCalificaciones.Corte2.Finalizado = finalizado
								}
								if InfoCalificaciones.Examen.IdEstado == IdEstado {
									InfoCalificaciones.Examen.Finalizado = finalizado
								}
								if InfoCalificaciones.Habilit.IdEstado == IdEstado {
									InfoCalificaciones.Habilit.Finalizado = finalizado
								}
								if InfoCalificaciones.Definitiva.IdEstado == IdEstado {
									InfoCalificaciones.Definitiva.Finalizado = finalizado
								}
							}

							EstadoRegistro := InfoCalificaciones.Corte1

							proyecto := findIdsbyId(proyectos, fmt.Sprintf("%v", asignatura.(map[string]interface{})["proyecto_academico_id"]))

							if fmt.Sprintf("%v", proyecto) != "map[]" {
								if InfoCalificaciones.Corte1.Finalizado {
									EstadoRegistro = InfoCalificaciones.Corte2
									if InfoCalificaciones.Corte2.Finalizado {
										EstadoRegistro = InfoCalificaciones.Examen
										if InfoCalificaciones.Examen.Finalizado {
											EstadoRegistro = InfoCalificaciones.Habilit
											if InfoCalificaciones.Habilit.Finalizado {
												EstadoRegistro = InfoCalificaciones.Definitiva
												if InfoCalificaciones.Definitiva.Finalizado {
													EstadoRegistro = InfoCalificaciones.Definitiva
												}
											}
										}
									}
								}

								resultados = append(resultados, map[string]interface{}{
									"Identificacion":     DocenteInfo[0]["Numero"],
									"Docente":            DocenteInfo[0]["TerceroId"].(map[string]interface{})["NombreCompleto"],
									"Codigo":             asignatura.(map[string]interface{})["codigo"],
									"Asignatura":         asignatura.(map[string]interface{})["nombre"],
									"Nivel":              proyecto["NivelFormacionId"].(map[string]interface{})["Nombre"],
									"Grupo":              asignatura.(map[string]interface{})["grupo"],
									"Inscritos":          asignatura.(map[string]interface{})["inscritos"],
									"Proyecto_Academico": proyecto["Nombre"],
									"Estado":             EstadoRegistro.Nombre,
									"Id":                 asignatura.(map[string]interface{})["_id"],
								})
							}

						} else {
							notOk = true
						}
					} else {
						notOk = true
					}

				}

				if !notOk {
					c.Ctx.Output.SetStatus(200)
					c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultados)
				} else {
					c.Ctx.Output.SetStatus(404)
					c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, "Error service GetEstadosRegistros: The request contains an incorrect parameter or no record exist")
				}
			} else {
				c.Ctx.Output.SetStatus(404)
				c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, "Error service GetEstadosRegistros: The request contains an incorrect parameter or no record exist")
			}
		} else {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, "Error service GetEstadosRegistros: The request contains an incorrect parameter or no record exist")
		}
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, "Error service GetEstadosRegistros: The request contains an incorrect parameter or no record exist")
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

	id_estudiante := c.Ctx.Input.Param(":id_estudiante")

	var EstudianteInformacion1 []interface{}
	var EstudianteInformacion2 []interface{}
	var proyecto []interface{}
	var calendario []interface{}
	var periodo map[string]interface{}
	var EspaciosAcademicos map[string]interface{}
	var InfoNota map[string]interface{}

	var notOk bool = false

	var resultados map[string]interface{}

	if InfoNotas, ok := EstadosRegistroIDs(); ok {

		if id_proyecto, ok := getProyectoFromEspacioAcademico_temporal(id_estudiante); ok {

			errEstudianteInformacion1 := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"info_complementaria_tercero?query=Activo:true,InfoComplementariaId.Id:93,TerceroId.Id:"+fmt.Sprintf("%v", id_estudiante), &EstudianteInformacion1)
			if errEstudianteInformacion1 == nil && fmt.Sprintf("%v", EstudianteInformacion1[0]) != "map[]" {
				errEstudianteInformacion2 := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"datos_identificacion?query=Activo:true,TerceroId:"+fmt.Sprintf("%v", id_estudiante)+"&sortby=Id&order=desc&limit=1", &EstudianteInformacion2) //,TipoDocumentoId.Id:3
				if errEstudianteInformacion2 == nil && fmt.Sprintf("%v", EstudianteInformacion2[0]) != "map[]" {
					errProyecto := request.GetJson("http://"+beego.AppConfig.String("ProyectoAcademicoService")+"proyecto_academico_institucion?query=Activo:true,Id:"+fmt.Sprintf("%v", id_proyecto), &proyecto)
					if errProyecto == nil && fmt.Sprintf("%v", proyecto[0]) != "map[]" {
						errCalendario := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario?query=Activo:true,Nivel:"+fmt.Sprintf("%v", proyecto[0].(map[string]interface{})["NivelFormacionId"].(map[string]interface{})["Id"]), &calendario)
						if errCalendario == nil && fmt.Sprintf("%v", calendario[0]) != "map[]" {
							errPeriodo := request.GetJson("http://"+beego.AppConfig.String("ParametroService")+"periodo/"+fmt.Sprintf("%v", calendario[0].(map[string]interface{})["PeriodoId"]), &periodo)
							if errPeriodo == nil && fmt.Sprintf("%v", periodo["Status"]) == "200" {
								errEspaciosAcademicos := request.GetJson("http://"+beego.AppConfig.String("EspaciosAcademicosService")+"espacio-academico-estudiantes?query=activo:true,periodo_id:"+fmt.Sprintf("%v", calendario[0].(map[string]interface{})["Id"])+",estudiante_id:"+fmt.Sprintf("%v", id_estudiante)+"&fields=_id&limit=0", &EspaciosAcademicos)
								if errEspaciosAcademicos == nil && fmt.Sprintf("%v", EspaciosAcademicos["Status"]) == "200" {

									var Asignaturas []interface{}

									for _, asignatura := range EspaciosAcademicos["Data"].([]interface{}) {
										var InfoAsignatura map[string]interface{}
										errInfoAsignatura := request.GetJson("http://"+beego.AppConfig.String("EspaciosAcademicosService")+"espacio-academico-estudiantes/"+fmt.Sprintf("%v", asignatura.(map[string]interface{})["_id"]), &InfoAsignatura)
										if errInfoAsignatura == nil && fmt.Sprintf("%v", InfoAsignatura["Status"]) == "200" {
											Asignaturas = append(Asignaturas, InfoAsignatura["Data"])
										} else {
											//algo bad
											notOk = true
										}
									}

									if !notOk {

										errInfoNota := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota?query=activo:true,estudiante_id:"+fmt.Sprintf("%v", id_estudiante)+"&fields=_id&limit=0", &InfoNota)
										if errInfoNota == nil && fmt.Sprintf("%v", InfoNota["Status"]) == "200" {

											var NotasDesagrupadas []interface{}

											for _, nota := range InfoNota["Data"].([]interface{}) {
												var InfoNotayReg map[string]interface{}
												errInfoNotayReg := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota/"+fmt.Sprintf("%v", nota.(map[string]interface{})["_id"]), &InfoNotayReg)
												if errInfoNotayReg == nil && fmt.Sprintf("%v", InfoNotayReg["Status"]) == "200" {
													NotasDesagrupadas = append(NotasDesagrupadas, InfoNotayReg["Data"])
												} else {
													//algo bad
													notOk = true
												}
											}

											if !notOk {

												var NotasAsignaturasEstudiante []interface{}

												for _, espAca := range Asignaturas {
													EspacioAcademico := espAca.(map[string]interface{})["espacio_academico_id"].(map[string]interface{})["_id"]
													for _, Nota := range NotasDesagrupadas {
														EspacioAcademicoNota := Nota.(map[string]interface{})["registro_id"].(map[string]interface{})["espacio_academico_id"]
														estadoRegistro := fmt.Sprintf("%v", Nota.(map[string]interface{})["registro_id"].(map[string]interface{})["estado_registro_id"])
														if EspacioAcademico == EspacioAcademicoNota && estadoRegistro == InfoNotas.Corte1.IdEstado {
															InfoNotas.Corte1.informacion = passNotaInfV2(Nota.(map[string]interface{}))
														}
														if EspacioAcademico == EspacioAcademicoNota && estadoRegistro == InfoNotas.Corte2.IdEstado {
															InfoNotas.Corte2.informacion = passNotaInfV2(Nota.(map[string]interface{}))
														}
														if EspacioAcademico == EspacioAcademicoNota && estadoRegistro == InfoNotas.Examen.IdEstado {
															InfoNotas.Examen.informacion = passNotaInfV2(Nota.(map[string]interface{}))
														}
														if EspacioAcademico == EspacioAcademicoNota && estadoRegistro == InfoNotas.Habilit.IdEstado {
															InfoNotas.Habilit.informacion = passNotaInfV2(Nota.(map[string]interface{}))
														}
														if EspacioAcademico == EspacioAcademicoNota && estadoRegistro == InfoNotas.Definitiva.IdEstado {
															InfoNotas.Definitiva.informacion = passNotaInfV2(Nota.(map[string]interface{}))
														}
													}
													NotasAsignaturasEstudiante = append(NotasAsignaturasEstudiante, map[string]interface{}{
														"Grupo":      espAca.(map[string]interface{})["espacio_academico_id"].(map[string]interface{})["grupo"],
														"Asignatura": espAca.(map[string]interface{})["espacio_academico_id"].(map[string]interface{})["nombre"],
														"Creditos":   espAca.(map[string]interface{})["espacio_academico_id"].(map[string]interface{})["creditos"],
														"Corte1":     InfoNotas.Corte1.informacion,
														"Corte2":     InfoNotas.Corte2.informacion,
														"Examen":     InfoNotas.Examen.informacion,
														"Habilit":    InfoNotas.Habilit.informacion,
														"Definitiva": InfoNotas.Definitiva.informacion,
														"Acumulado":  calculoAcumuladoNotas(InfoNotas),
													})
												}

												resultados = map[string]interface{}{
													"Nombre":              EstudianteInformacion2[0].(map[string]interface{})["TerceroId"].(map[string]interface{})["NombreCompleto"],
													"Identificacion":      EstudianteInformacion2[0].(map[string]interface{})["Numero"],
													"Codigo":              EstudianteInformacion1[0].(map[string]interface{})["Dato"],
													"Codigo_programa":     proyecto[0].(map[string]interface{})["Codigo"],
													"Nombre_programa":     proyecto[0].(map[string]interface{})["Nombre"],
													"Promedio":            "falta",
													"Periodo":             periodo["Data"].(map[string]interface{})["Nombre"],
													"Espacios_academicos": NotasAsignaturasEstudiante,
												}

												c.Ctx.Output.SetStatus(200)
												c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultados)

											} else {
												c.Ctx.Output.SetStatus(404)
												c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, "notok2 Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
											}
										} else {
											logs.Error(errInfoNota)
											c.Ctx.Output.SetStatus(404)
											c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, "errInfoNota Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
										}
									} else {
										c.Ctx.Output.SetStatus(404)
										c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, "notok1 Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
									}
								} else {
									logs.Error(errEspaciosAcademicos)
									c.Ctx.Output.SetStatus(404)
									c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, "errEspaciosAcademicos Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
								}
							} else {
								logs.Error(errPeriodo)
								c.Ctx.Output.SetStatus(404)
								c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, "errPeriodo Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
							}
						} else {
							logs.Error(errCalendario)
							c.Ctx.Output.SetStatus(404)
							c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, "errCalendario Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
						}
					} else {
						logs.Error(errProyecto)
						c.Ctx.Output.SetStatus(404)
						c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, "errProyecto Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
					}
				} else {
					logs.Error(errEstudianteInformacion2)
					c.Ctx.Output.SetStatus(404)
					c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, "errEstudianteInformacion2 Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
				}
			} else {
				logs.Error(errEstudianteInformacion1)
				c.Ctx.Output.SetStatus(404)
				c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, "errEstudianteInformacion1 Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
			}
		} else {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, "EstadosRegistroIDs Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
		}
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, "EstadosRegistroIDs Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
	}

	c.ServeJSON()
}

func getProyectoFromEspacioAcademico_temporal(id_estudiante string) (string, bool) {
	ok := false
	proyecto_academico_id := "0"
	var espAcaEst map[string]interface{}
	erresAcaEst := request.GetJson("http://"+beego.AppConfig.String("EspaciosAcademicosService")+"espacio-academico-estudiantes?query=activo:true,estudiante_id:"+fmt.Sprintf("%v", id_estudiante)+"&fields=espacio_academico_id&limit=1", &espAcaEst)
	if erresAcaEst == nil && fmt.Sprintf("%v", espAcaEst["Status"]) == "200" {
		id := espAcaEst["Data"].([]interface{})[0].(map[string]interface{})["espacio_academico_id"]
		var espAca map[string]interface{}
		erresAca := request.GetJson("http://"+beego.AppConfig.String("EspaciosAcademicosService")+"espacio-academico?query=activo:true,_id:"+fmt.Sprintf("%v", id)+"&fields=proyecto_academico_id&limit=1", &espAca)
		if erresAca == nil && fmt.Sprintf("%v", espAca["Status"]) == "200" {
			proyecto_academico_id = fmt.Sprintf("%v", espAca["Data"].([]interface{})[0].(map[string]interface{})["proyecto_academico_id"])
			ok = true
		}
	}
	return proyecto_academico_id, ok
}

type TipoEstado struct {
	IdEstado         string
	Nombre           string
	Existe           bool
	IdRegistroNota   string
	Finalizado       bool
	EditExtemporaneo bool
	informacion      map[string]interface{}
}

type EstadosRegistro struct {
	Corte1     TipoEstado
	Corte2     TipoEstado
	Examen     TipoEstado
	Habilit    TipoEstado
	Definitiva TipoEstado
}

func EstadosRegistroIDs() (EstadosRegistro, bool) {

	EstadosRegistros := EstadosRegistro{
		Corte1:     TipoEstado{Nombre: "PRIMER CORTE"},
		Corte2:     TipoEstado{Nombre: "SEGUNDO CORTE"},
		Examen:     TipoEstado{Nombre: "EXAMEN FINAL"},
		Habilit:    TipoEstado{Nombre: "HABILITACIONES"},
		Definitiva: TipoEstado{Nombre: "DEFINITIVA"},
	}

	var EstadosRegistroApi map[string]interface{}
	errEstadosRegistroApi := request.GetJson("http://"+beego.AppConfig.String("ParametroService")+"parametro?query=TipoParametroId:52&fields=Id,Nombre&limit=0", &EstadosRegistroApi)
	if errEstadosRegistroApi == nil && fmt.Sprintf("%v", EstadosRegistroApi["Status"]) == "200" && fmt.Sprintf("%v", EstadosRegistroApi["Data"]) != "[map[]]" {
		for _, EstReg := range EstadosRegistroApi["Data"].([]interface{}) {
			if fmt.Sprintf("%v", EstReg.(map[string]interface{})["Nombre"]) == EstadosRegistros.Corte1.Nombre {
				EstadosRegistros.Corte1.IdEstado = fmt.Sprintf("%v", EstReg.(map[string]interface{})["Id"])
				EstadosRegistros.Corte1.Existe = false
				EstadosRegistros.Corte1.Finalizado = false
			}
			if fmt.Sprintf("%v", EstReg.(map[string]interface{})["Nombre"]) == EstadosRegistros.Corte2.Nombre {
				EstadosRegistros.Corte2.IdEstado = fmt.Sprintf("%v", EstReg.(map[string]interface{})["Id"])
				EstadosRegistros.Corte2.Existe = false
				EstadosRegistros.Corte2.Finalizado = false
			}
			if fmt.Sprintf("%v", EstReg.(map[string]interface{})["Nombre"]) == EstadosRegistros.Examen.Nombre {
				EstadosRegistros.Examen.IdEstado = fmt.Sprintf("%v", EstReg.(map[string]interface{})["Id"])
				EstadosRegistros.Examen.Existe = false
				EstadosRegistros.Examen.Finalizado = false
			}
			if fmt.Sprintf("%v", EstReg.(map[string]interface{})["Nombre"]) == EstadosRegistros.Habilit.Nombre {
				EstadosRegistros.Habilit.IdEstado = fmt.Sprintf("%v", EstReg.(map[string]interface{})["Id"])
				EstadosRegistros.Habilit.Existe = false
				EstadosRegistros.Habilit.Finalizado = false
			}
			if fmt.Sprintf("%v", EstReg.(map[string]interface{})["Nombre"]) == EstadosRegistros.Definitiva.Nombre {
				EstadosRegistros.Definitiva.IdEstado = fmt.Sprintf("%v", EstReg.(map[string]interface{})["Id"])
				EstadosRegistros.Definitiva.Existe = false
				EstadosRegistros.Definitiva.Finalizado = false
			}
		}
		return EstadosRegistros, true
	} else {
		return EstadosRegistro{}, false
	}
}

func calculoNotasPorCortes(Notas map[string]interface{}) map[string]interface{} {

	var calculo = 0.0

	for _, nota := range Notas["valor_nota"].([]interface{}) {
		perc := nota.(map[string]interface{})["perc"].(float64)
		value := nota.(map[string]interface{})["value"].(float64)
		calc := perc / 100.0 * value
		calculo += calc
	}

	Notas["nota_definitiva"] = calculo

	return Notas
}

func calculoDefinitiva(NotasEst interface{}) float64 {

	var calculo = 0.0

	n1 := NotasEst.(map[string]interface{})["Corte1"].(map[string]interface{})["data"].(map[string]interface{})["nota_definitiva"].(float64)
	n2 := NotasEst.(map[string]interface{})["Corte2"].(map[string]interface{})["data"].(map[string]interface{})["nota_definitiva"].(float64)
	n3 := NotasEst.(map[string]interface{})["Examen"].(map[string]interface{})["data"].(map[string]interface{})["nota_definitiva"].(float64)
	n4 := NotasEst.(map[string]interface{})["Habilit"].(map[string]interface{})["data"].(map[string]interface{})["nota_definitiva"].(float64)

	calculo = n1 + n2 + n3 + n4

	Definitiva := math.Round(calculo*10) / 10

	return Definitiva
}

func calculoAcumuladoNotas(calif EstadosRegistro) float64 {

	calculo := 0.0

	n1 := calif.Corte1.informacion["data"].(map[string]interface{})["nota_definitiva"].(float64)
	n2 := calif.Corte2.informacion["data"].(map[string]interface{})["nota_definitiva"].(float64)
	n3 := calif.Examen.informacion["data"].(map[string]interface{})["nota_definitiva"].(float64)
	n4 := calif.Habilit.informacion["data"].(map[string]interface{})["nota_definitiva"].(float64)

	calculo = n1 + n2 + n3 + n4

	Acumulado := math.Floor(calculo*100) / 100

	return Acumulado
}

func passNotaInf(N map[string]interface{}) map[string]interface{} {
	n := map[string]interface{}{
		"id": N["Data"].([]interface{})[0].(map[string]interface{})["_id"],
		"data": map[string]interface{}{
			"valor_nota":          N["Data"].([]interface{})[0].(map[string]interface{})["valor_nota"],
			"nota_definitiva":     N["Data"].([]interface{})[0].(map[string]interface{})["nota_definitiva"],
			"fallas":              N["Data"].([]interface{})[0].(map[string]interface{})["fallas"],
			"observacion_nota_id": N["Data"].([]interface{})[0].(map[string]interface{})["observacion_nota_id"],
		},
	}
	return n
}

func passNotaInfV2(N map[string]interface{}) map[string]interface{} {
	n := map[string]interface{}{
		"data": map[string]interface{}{
			"valor_nota":          N["valor_nota"],
			"nota_definitiva":     N["nota_definitiva"],
			"fallas":              N["fallas"],
			"observacion_nota_id": N["observacion_nota_id"],
		},
	}
	return n
}

func passNotaEmpty() map[string]interface{} {
	n := map[string]interface{}{
		"id": "",
		"data": map[string]interface{}{
			"valor_nota":          map[string]interface{}{},
			"nota_definitiva":     0,
			"fallas":              0,
			"observacion_nota_id": 0,
		},
	}
	return n
}

func passPorcentajeEmpty(reg string) map[string]interface{} {
	regI, _ := strconv.Atoi(reg)
	p := map[string]interface{}{
		"id":               "",
		"estadoRegistro":   regI,
		"fields":           map[string]interface{}{},
		"editExtemporaneo": false,
		"finalizado":       false,
	}
	return p
}

func findNamebyId(list []interface{}, id string) string {
	for _, item := range list {
		if fmt.Sprintf("%v", item.(map[string]interface{})["Id"]) == id {
			return fmt.Sprintf("%v", item.(map[string]interface{})["Nombre"])
		}
	}
	return ""
}

func findIdsbyId(list []interface{}, id string) map[string]interface{} {
	for _, item := range list {
		if fmt.Sprintf("%v", item.(map[string]interface{})["Id"]) == id {
			return item.(map[string]interface{})
		}
	}
	return map[string]interface{}{}
}

func validatePutPorcentajes(p map[string]interface{}) bool {
	valid := false

	if Accion, ok := p["Accion"]; ok {
		if reflect.TypeOf(Accion).Kind() == reflect.String {
			if Estado_registro, ok := p["Estado_registro"]; ok {
				if reflect.TypeOf(Estado_registro).Kind() == reflect.Float64 {
					if PorcentajesNotas, ok := p["PorcentajesNotas"]; ok {
						if reflect.TypeOf(PorcentajesNotas).Kind() == reflect.Slice {
							for _, r := range p["PorcentajesNotas"].([]interface{}) {
								if editExtemporaneo, ok := r.(map[string]interface{})["editExtemporaneo"]; ok {
									if reflect.TypeOf(editExtemporaneo).Kind() == reflect.Bool {
										if estadoRegistro, ok := r.(map[string]interface{})["estadoRegistro"]; ok {
											if reflect.TypeOf(estadoRegistro).Kind() == reflect.Float64 {
												if fields, ok := r.(map[string]interface{})["fields"]; ok {
													if reflect.TypeOf(fields).Kind() == reflect.Map {
														if finalizado, ok := r.(map[string]interface{})["finalizado"]; ok {
															if reflect.TypeOf(finalizado).Kind() == reflect.Bool {
																if id, ok := r.(map[string]interface{})["id"]; ok {
																	if reflect.TypeOf(id).Kind() == reflect.String {
																		if editporTiempo, ok := r.(map[string]interface{})["editporTiempo"]; ok {
																			if reflect.TypeOf(editporTiempo).Kind() == reflect.Bool {
																				valid = true
																			} else {
																				valid = false
																				break
																			}
																		} else {
																			valid = false
																			break
																		}
																	} else {
																		valid = false
																		break
																	}
																} else {
																	valid = false
																	break
																}
															} else {
																valid = false
																break
															}
														} else {
															valid = false
															break
														}
													} else {
														valid = false
														break
													}
												} else {
													valid = false
													break
												}
											} else {
												valid = false
												break
											}
										} else {
											valid = false
											break
										}
									} else {
										valid = false
										break
									}
								} else {
									valid = false
									break
								}
							}
						} else {
							valid = false
						}
					} else {
						valid = false
					}
				} else {
					valid = false
				}
			} else {
				valid = false
			}
			if Accion == "Crear" {
				if Info, ok := p["Info"]; ok {
					if reflect.TypeOf(Info).Kind() == reflect.Map {
						if nombre, ok := Info.(map[string]interface{})["nombre"]; ok {
							if reflect.TypeOf(nombre).Kind() == reflect.String {
								if codigo, ok := Info.(map[string]interface{})["codigo"]; ok {
									if reflect.TypeOf(codigo).Kind() == reflect.String {
										if periodo, ok := Info.(map[string]interface{})["periodo"]; ok {
											if reflect.TypeOf(periodo).Kind() == reflect.Float64 {
												if nivel, ok := Info.(map[string]interface{})["nivel"]; ok {
													if reflect.TypeOf(nivel).Kind() == reflect.Float64 {
														if espacio_academico, ok := Info.(map[string]interface{})["espacio_academico"]; ok {
															if reflect.TypeOf(espacio_academico).Kind() == reflect.String {
																valid = true
															} else {
																valid = false
															}
														} else {
															valid = false
														}
													} else {
														valid = false
													}
												} else {
													valid = false
												}
											} else {
												valid = false
											}
										} else {
											valid = false
										}
									} else {
										valid = false
									}
								} else {
									valid = false
								}
							} else {
								valid = false
							}
						} else {
							valid = false
						}
					} else {
						valid = false
					}
				} else {
					valid = false
				}
			}
		} else {
			valid = false
		}
	} else {
		valid = false
	}

	return valid
}

func validatePutNotasEstudiantes(n map[string]interface{}) bool {
	valid := false

	if Accion, ok := n["Accion"]; ok {
		if reflect.TypeOf(Accion).Kind() == reflect.String {
			if Estado_registro, ok := n["Estado_registro"]; ok {
				if reflect.TypeOf(Estado_registro).Kind() == reflect.Float64 {
					if Espacio_academico, ok := n["Espacio_academico"]; ok {
						if reflect.TypeOf(Espacio_academico).Kind() == reflect.String {
							if Nombre, ok := n["Nombre"]; ok {
								if reflect.TypeOf(Nombre).Kind() == reflect.String {
									if Periodo, ok := n["Periodo"]; ok {
										if reflect.TypeOf(Periodo).Kind() == reflect.Float64 {
											if CalificacionesEstudiantes, ok := n["CalificacionesEstudiantes"]; ok {
												if reflect.TypeOf(CalificacionesEstudiantes).Kind() == reflect.Slice {
													for _, e := range CalificacionesEstudiantes.([]interface{}) {
														if Id, ok := e.(map[string]interface{})["Id"]; ok {
															if reflect.TypeOf(Id).Kind() == reflect.Float64 {
																if Corte1, ok := e.(map[string]interface{})["Corte1"]; ok {
																	if reflect.TypeOf(Corte1).Kind() == reflect.Map {
																		if id, ok := Corte1.(map[string]interface{})["id"]; ok {
																			if reflect.TypeOf(id).Kind() == reflect.String {
																				if data, ok := Corte1.(map[string]interface{})["data"]; ok {
																					if reflect.TypeOf(data).Kind() == reflect.Map {
																						valid = true
																					} else {
																						valid = false
																						break
																					}
																				} else {
																					valid = false
																					break
																				}
																			} else {
																				valid = false
																				break
																			}
																		} else {
																			valid = false
																			break
																		}
																	} else {
																		valid = false
																		break
																	}
																}
																if Corte2, ok := e.(map[string]interface{})["Corte2"]; ok {
																	if reflect.TypeOf(Corte2).Kind() == reflect.Map {
																		if id, ok := Corte2.(map[string]interface{})["id"]; ok {
																			if reflect.TypeOf(id).Kind() == reflect.String {
																				if data, ok := Corte2.(map[string]interface{})["data"]; ok {
																					if reflect.TypeOf(data).Kind() == reflect.Map {
																						valid = true
																					} else {
																						valid = false
																						break
																					}
																				} else {
																					valid = false
																					break
																				}
																			} else {
																				valid = false
																				break
																			}
																		} else {
																			valid = false
																			break
																		}
																	} else {
																		valid = false
																		break
																	}
																}

																if Examen, ok := e.(map[string]interface{})["Examen"]; ok {
																	if reflect.TypeOf(Examen).Kind() == reflect.Map {
																		if id, ok := Examen.(map[string]interface{})["id"]; ok {
																			if reflect.TypeOf(id).Kind() == reflect.String {
																				if data, ok := Examen.(map[string]interface{})["data"]; ok {
																					if reflect.TypeOf(data).Kind() == reflect.Map {
																						valid = true
																					} else {
																						valid = false
																						break
																					}
																				} else {
																					valid = false
																					break
																				}
																			} else {
																				valid = false
																				break
																			}
																		} else {
																			valid = false
																			break
																		}
																	} else {
																		valid = false
																		break
																	}
																}

																if Habilit, ok := e.(map[string]interface{})["Habilit"]; ok {
																	if reflect.TypeOf(Habilit).Kind() == reflect.Map {
																		if id, ok := Habilit.(map[string]interface{})["id"]; ok {
																			if reflect.TypeOf(id).Kind() == reflect.String {
																				if data, ok := Habilit.(map[string]interface{})["data"]; ok {
																					if reflect.TypeOf(data).Kind() == reflect.Map {
																						valid = true
																					} else {
																						valid = false
																						break
																					}
																				} else {
																					valid = false
																					break
																				}
																			} else {
																				valid = false
																				break
																			}
																		} else {
																			valid = false
																			break
																		}
																	} else {
																		valid = false
																		break
																	}
																}

																if Definitiva, ok := e.(map[string]interface{})["Definitiva"]; ok {
																	if reflect.TypeOf(Definitiva).Kind() == reflect.Map {
																		if id, ok := Definitiva.(map[string]interface{})["id"]; ok {
																			if reflect.TypeOf(id).Kind() == reflect.String {
																				if data, ok := Definitiva.(map[string]interface{})["data"]; ok {
																					if reflect.TypeOf(data).Kind() == reflect.Map {
																						valid = true
																					} else {
																						valid = false
																						break
																					}
																				} else {
																					valid = false
																					break
																				}
																			} else {
																				valid = false
																				break
																			}
																		} else {
																			valid = false
																			break
																		}
																	} else {
																		valid = false
																		break
																	}
																}

															} else {
																valid = false
																break
															}
														} else {
															valid = false
															break
														}
													}
												} else {
													valid = false
												}
											} else {
												valid = false
											}
										} else {
											valid = false
										}
									} else {
										valid = false
									}
								} else {
									valid = false
								}
							} else {
								valid = false
							}
						} else {
							valid = false
						}
					} else {
						valid = false
					}
				} else {
					valid = false
				}
			} else {
				valid = false
			}
		} else {
			valid = false
		}
	} else {
		valid = false
	}

	return valid
}
