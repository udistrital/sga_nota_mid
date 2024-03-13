package services

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/prometheus/common/log"
	"github.com/udistrital/sga_nota_mid/helpers"
	"github.com/udistrital/utils_oas/request"
)

func GetEspaciosAcademicosDocente(idDocente string) ([]interface{}, error) {
	espaciosAcademicos, err := helpers.ObtenerEspaciosAcademicos(idDocente)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	proyectos, err := helpers.ObtenerProyectos()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	calendarios, err := helpers.ObtenerCalendarios()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	periodos, err := helpers.ObtenerPeriodos()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var resultados []interface{}

	if helpers.VerificarNil(espaciosAcademicos["Data"]) {
		return nil, errors.New("error en espaciosAcademicos: la clave 'Data' está vacía o no existe")
	}

	if helpers.VerificarNil(calendarios) {
		return nil, errors.New("error en calendarios: está vacía o no existe")
	}

	if helpers.VerificarNil(proyectos) {
		return nil, errors.New("error en proyectos: está vacía o no existe")
	}

	if helpers.VerificarNil(periodos["Data"]) {
		return nil, errors.New("error en periodos: la clave ['data'] está vacía o no existe")
	}

	for _, espacioAcademicoRegistro := range espaciosAcademicos["Data"].([]interface{}) {
		registro := espacioAcademicoRegistro.(map[string]interface{})

		calendario := helpers.FindIdsbyId(calendarios, fmt.Sprintf("%v", registro["periodo_id"]))
		if len(calendario) == 0 {
			continue
		}

		proyecto := helpers.FindIdsbyId(proyectos, fmt.Sprintf("%v", registro["proyecto_academico_id"]))
		if len(proyecto) == 0 {
			continue
		}

		periodoNombre := helpers.FindNamebyId(periodos["Data"].([]interface{}), fmt.Sprintf("%v", calendario["PeriodoId"]))

		resultado := map[string]interface{}{
			"Nivel":              proyecto["NivelFormacionId"].(map[string]interface{})["Nombre"],
			"Nivel_id":           proyecto["NivelFormacionId"].(map[string]interface{})["Id"],
			"Codigo":             registro["codigo"],
			"Asignatura":         registro["nombre"],
			"Periodo":            periodoNombre,
			"PeriodoId":          registro["periodo_id"],
			"Grupo":              registro["grupo"],
			"Inscritos":          registro["inscritos"],
			"Proyecto_Academico": proyecto["Nombre"],
			"AsignaturaId":       registro["_id"],
		}

		resultados = append(resultados, resultado)
	}

	if len(resultados) == 0 {
		return nil, errors.New("no se encontraron datos")
	}

	return resultados, nil
}

func GetModificacionExtemporanea(idAsignatura string) (map[string]interface{}, error) {
	var RegistroAsignatura map[string]interface{}
	url := fmt.Sprintf("http://%sregistro?query=activo:true,espacio_academico_id:%v&fields=estado_registro_id,modificacion_extemporanea&limit=0", beego.AppConfig.String("CalificacionesService"), idAsignatura)
	err := request.GetJson(url, &RegistroAsignatura)
	return RegistroAsignatura, err
}

func GetDatosDocenteAsignatura(idAsignatura string) ([]interface{}, error) {
	resultado := []interface{}{}
	var EspacioAcademicoRegistro map[string]interface{}
	var DocenteInfo []map[string]interface{}
	var proyecto []interface{}
	var calendario []interface{}
	var periodo map[string]interface{}

	errEspacioAcademicoRegistro := request.GetJson("http://"+beego.AppConfig.String("EspaciosAcademicosService")+"espacio-academico/"+fmt.Sprintf("%v", idAsignatura), &EspacioAcademicoRegistro)
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
						return resultado, nil
					} else {
						log.Error(errPeriodo)
						return nil, errors.New("error en GetDatosDocenteAsignatura: periodo está vacía o no existe")
					}
				} else {
					logs.Error(errCalendario)
					return nil, errors.New("error en GetDatosDocenteAsignatura: calendario está vacía o no existe")
				}
			} else {
				logs.Error(errProyecto)
				return nil, errors.New("error en GetDatosDocenteAsignatura: proyecto está vacía o no existe")
			}
		} else {
			logs.Error(errDocenteInfo)
			return nil, errors.New("error en GetDatosDocenteAsignatura: docente está vacía o no existe")
		}
	} else {
		logs.Error(errEspacioAcademicoRegistro)
		return nil, errors.New("error en GetDatosDocenteAsignatura: espacioAcademicoRegistro está vacía o no existe")
	}
}

func GetPorcentajesAsignatura(idAsignatura string, idPeriodo string) ([]interface{}, error) {

	if InfoPorcentajes, ok := helpers.EstadosRegistroIDs(); ok {

		resultados := []interface{}{}

		var RegistroAsignatura map[string]interface{}
		errRegistroAsignatura := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"registro?query=activo:true,espacio_academico_id:"+fmt.Sprintf("%v", idAsignatura)+",periodo_id:"+fmt.Sprintf("%v", idPeriodo), &RegistroAsignatura)
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
				resultados = append(resultados, helpers.PassPorcentajeEmpty(InfoPorcentajes.Corte1.IdEstado))
			}
			if !InfoPorcentajes.Corte2.Existe {
				resultados = append(resultados, helpers.PassPorcentajeEmpty(InfoPorcentajes.Corte2.IdEstado))
			}
			if !InfoPorcentajes.Examen.Existe {
				resultados = append(resultados, helpers.PassPorcentajeEmpty(InfoPorcentajes.Examen.IdEstado))
			}
			if !InfoPorcentajes.Habilit.Existe {
				resultados = append(resultados, helpers.PassPorcentajeEmpty(InfoPorcentajes.Habilit.IdEstado))
			}
			if !InfoPorcentajes.Definitiva.Existe {
				resultados = append(resultados, helpers.PassPorcentajeEmpty(InfoPorcentajes.Definitiva.IdEstado))
			}

			return resultados, nil

		} else {
			logs.Error(errRegistroAsignatura)
			return nil, errors.New("error en GetPorcentajesAsignatura: registroAsignatura está vacía o no existe")
		}
	} else {
		return nil, errors.New("error en GetPorcentajesAsignatura: EstadoDeRegistro está vacía o no existe")
	}
}

func PutPorcentajeAsignatura(data []byte) (interface{}, error) {
	var inputData map[string]interface{}

	if err := json.Unmarshal(data, &inputData); err == nil {

		valido := helpers.ValidatePutPorcentajes(inputData)

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
					return nil, errors.New("error en PutPorcentajesAsignatura: crearSalioMal")
				} else {
					return crearRegistrosReporte, nil
				}
			} else if guardarRegistros {
				if guardarSalioMal {
					return nil, errors.New("error en PutPorcentajesAsignatura: guardarSalioMal")
				} else {
					return guardarRegistroReporte, nil
				}
			} else {
				return nil, errors.New("error en PutPorcentajesAsignatura: guardarSalioMal")
			}
		} else {
			return nil, errors.New("error en PutPorcentajesAsignatura: No es valido")
		}
	} else {
		logs.Error(err)
		return nil, errors.New(err.Error())
	}
}

func GetCapturaNotas(idEspacioAcademico string, idPeriodo string) (interface{}, error) {
	var resultado map[string]interface{}
	datos := []interface{}{}

	var EspaciosAcademicosEstudiantes map[string]interface{}
	var RegistroCalificacion map[string]interface{}
	var EstudianteInformacion []interface{}

	if InformacionCalificaciones, ok := helpers.EstadosRegistroIDs(); ok {

		errRegistroCalificacion := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"registro?query=activo:true,periodo_id:"+fmt.Sprintf("%v", idPeriodo)+",espacio_academico_id:"+fmt.Sprintf("%v", idEspacioAcademico), &RegistroCalificacion)
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

			errEspaciosAcademicosEstudiantes := request.GetJson("http://"+beego.AppConfig.String("EspaciosAcademicosService")+"espacio-academico-estudiantes?query=activo:true,espacio_academico_id:"+fmt.Sprintf("%v", idEspacioAcademico)+",periodo_id:"+fmt.Sprintf("%v", idPeriodo), &EspaciosAcademicosEstudiantes)
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
								InformacionCalificaciones.Corte1.Informacion = helpers.PassNotaInf(InfoNota)
							} else {
								InformacionCalificaciones.Corte1.Informacion = helpers.PassNotaEmpty()
							}
						} else {
							InformacionCalificaciones.Corte1.Informacion = helpers.PassNotaEmpty()
						}

						if InformacionCalificaciones.Corte2.Existe {
							var InfoNota map[string]interface{}
							errInfoNota := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota?query=activo:true,registro_id:"+InformacionCalificaciones.Corte2.IdRegistroNota+",estudiante_id:"+fmt.Sprintf("%v", id_estudiante), &InfoNota)
							if errInfoNota == nil && fmt.Sprintf("%v", InfoNota["Status"]) == "200" {
								InformacionCalificaciones.Corte2.Informacion = helpers.PassNotaInf(InfoNota)
							} else {
								InformacionCalificaciones.Corte2.Informacion = helpers.PassNotaEmpty()
							}
						} else {
							InformacionCalificaciones.Corte2.Informacion = helpers.PassNotaEmpty()
						}

						if InformacionCalificaciones.Examen.Existe {
							var InfoNota map[string]interface{}
							errInfoNota := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota?query=activo:true,registro_id:"+InformacionCalificaciones.Examen.IdRegistroNota+",estudiante_id:"+fmt.Sprintf("%v", id_estudiante), &InfoNota)
							if errInfoNota == nil && fmt.Sprintf("%v", InfoNota["Status"]) == "200" {
								InformacionCalificaciones.Examen.Informacion = helpers.PassNotaInf(InfoNota)
							} else {
								InformacionCalificaciones.Examen.Informacion = helpers.PassNotaEmpty()
							}
						} else {
							InformacionCalificaciones.Examen.Informacion = helpers.PassNotaEmpty()
						}

						if InformacionCalificaciones.Habilit.Existe {
							var InfoNota map[string]interface{}
							errInfoNota := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota?query=activo:true,registro_id:"+InformacionCalificaciones.Habilit.IdRegistroNota+",estudiante_id:"+fmt.Sprintf("%v", id_estudiante), &InfoNota)
							if errInfoNota == nil && fmt.Sprintf("%v", InfoNota["Status"]) == "200" {
								InformacionCalificaciones.Habilit.Informacion = helpers.PassNotaInf(InfoNota)
							} else {
								InformacionCalificaciones.Habilit.Informacion = helpers.PassNotaEmpty()
							}
						} else {
							InformacionCalificaciones.Habilit.Informacion = helpers.PassNotaEmpty()
						}

						if InformacionCalificaciones.Definitiva.Existe {
							var InfoNota map[string]interface{}
							errInfoNota := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota?query=activo:true,registro_id:"+InformacionCalificaciones.Definitiva.IdRegistroNota+",estudiante_id:"+fmt.Sprintf("%v", id_estudiante), &InfoNota)
							if errInfoNota == nil && fmt.Sprintf("%v", InfoNota["Status"]) == "200" {
								InformacionCalificaciones.Definitiva.Informacion = helpers.PassNotaInf(InfoNota)
							} else {
								InformacionCalificaciones.Definitiva.Informacion = helpers.PassNotaEmpty()
							}
						} else {
							InformacionCalificaciones.Definitiva.Informacion = helpers.PassNotaEmpty()
						}

						datos = append(datos, map[string]interface{}{
							"Id":         id_estudiante,
							"Codigo":     Codigo,
							"Nombre":     fmt.Sprintf("%v", Nombre1) + " " + fmt.Sprintf("%v", Nombre2),
							"Apellido":   fmt.Sprintf("%v", Apellido1) + " " + fmt.Sprintf("%v", Apellido2),
							"Corte1":     InformacionCalificaciones.Corte1.Informacion,
							"Corte2":     InformacionCalificaciones.Corte2.Informacion,
							"Examen":     InformacionCalificaciones.Examen.Informacion,
							"Habilit":    InformacionCalificaciones.Habilit.Informacion,
							"Definitiva": InformacionCalificaciones.Definitiva.Informacion,
							"Acumulado":  helpers.CalculoAcumuladoNotas(InformacionCalificaciones),
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
				return resultado, nil

			} else {
				logs.Error(errEspaciosAcademicosEstudiantes)
				return nil, errors.New("error al consultar espacios academicos o la informacion no existe")
			}
		} else {
			logs.Error(errRegistroCalificacion)
			return nil, errors.New("error al consultar registro de calificaciones o la informacion no existe")
		}
	} else {
		return nil, errors.New("error al cargar informacion de calificaciones o la informacion no existe")
	}
}

func PutCapturaNotas(data []byte) (interface{}, error) {
	var inputData map[string]interface{}

	if err := json.Unmarshal(data, &inputData); err == nil {

		valido := helpers.ValidatePutNotasEstudiantes(inputData)

		if valido {

			if InfoCalificaciones, ok := helpers.EstadosRegistroIDs(); ok {

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

								nota_json = helpers.CalculoNotasPorCortes(nota_json.(map[string]interface{}))
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

								nota_json = helpers.CalculoNotasPorCortes(nota_json.(map[string]interface{}))
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

								nota_json = helpers.CalculoNotasPorCortes(nota_json.(map[string]interface{}))
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

								nota_json = helpers.CalculoNotasPorCortes(nota_json.(map[string]interface{}))
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
								nota_json = helpers.CalculoNotasPorCortes(nota_json.(map[string]interface{}))
								nota_json.(map[string]interface{})["fallas"] = fallas
								nota_json.(map[string]interface{})["observacion_nota_id"] = Observ
								if fmt.Sprintf("%v", ObservCod) == "3" {
									nota_json.(map[string]interface{})["nota_definitiva"] = 0
									nota_json.(map[string]interface{})["aprobado"] = false
								} else {
									def = helpers.CalculoDefinitiva(CalificacionEstudiante)
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
							return nil, errors.New("error del servicio PutCapturaNotas: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
						} else {
							return crearNotasReporte, nil
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
							return nil, errors.New("error del servicio PutCapturaNotas: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
						} else {
							return guardarNotasReporte, nil
						}
					}
				} else {
					logs.Error(errInfoRegistro)
					return nil, errors.New("error del servicio PutCapturaNotas: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
				}
			} else {
				return nil, errors.New("error del servicio PutCapturaNotas: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
			}
		} else {
			return nil, errors.New("error del servicio PutCapturaNotas: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
		}
	} else {
		logs.Error(err)
		return nil, errors.New("error del servicio PutCapturaNotas: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
	}
	return nil, errors.New("error del servicio PutCapturaNotas: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
}

func GetEstadosRegistros(idPeriodo string) (interface{}, error) {
	resultados := []interface{}{}

	var EspaciosAcademicosRegistros map[string]interface{}
	var proyectos []interface{}

	var notOk bool = false

	if InfoCalificaciones, ok := helpers.EstadosRegistroIDs(); ok {
		errEspaciosAcademicosRegistros := request.GetJson("http://"+beego.AppConfig.String("EspaciosAcademicosService")+"espacio-academico?query=activo:true,periodo_id:"+fmt.Sprintf("%v", idPeriodo)+"&limit=0", &EspaciosAcademicosRegistros)
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
						errInfoRegistro := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"registro?query=activo:true,periodo_id:"+fmt.Sprintf("%v", idPeriodo)+",espacio_academico_id:"+fmt.Sprintf("%v", espacioAcademico), &InfoRegistro)
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

							proyecto := helpers.FindIdsbyId(proyectos, fmt.Sprintf("%v", asignatura.(map[string]interface{})["proyecto_academico_id"]))

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
							return nil, errors.New("error del servicio GetEstadosRegistros: [info registros] La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
						}
					} else {
						return nil, errors.New("error del servicio GetEstadosRegistros: [docente info] La solicitud contiene un tipo de dato incorrecto o un parámetro inválido" + errDocenteInfo.Error())
					}

				}

				if !notOk {
					return resultados, nil
				} else {
					return nil, errors.New("error del servicio GetEstadosRegistros: [resultados] La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
				}
			} else {
				return nil, errors.New("error del servicio GetEstadosRegistros: [proyectos] La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
			}
		} else {
			return nil, errors.New("error del servicio GetEstadosRegistros: [espacios academicos] La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
		}
	} else {
		return nil, errors.New("error del servicio GetEstadosRegistros: [info calificaciones] La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
	}
}

func GetDatosEstudianteNotas(idEstudiante string) (interface{}, error) {
	var EstudianteInformacion1 []interface{}
	var EstudianteInformacion2 []interface{}
	var proyecto []interface{}
	var calendario []interface{}
	var periodo map[string]interface{}
	var EspaciosAcademicos map[string]interface{}
	var InfoNota map[string]interface{}

	var notOk bool = false

	var resultados map[string]interface{}

	if InfoNotas, ok := helpers.EstadosRegistroIDs(); ok {

		if id_proyecto, ok := helpers.GetProyectoFromEspacioAcademico_temporal(idEstudiante); ok {

			errEstudianteInformacion1 := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"info_complementaria_tercero?query=Activo:true,InfoComplementariaId.Id:93,TerceroId.Id:"+fmt.Sprintf("%v", idEstudiante), &EstudianteInformacion1)
			if errEstudianteInformacion1 == nil && fmt.Sprintf("%v", EstudianteInformacion1[0]) != "map[]" {
				errEstudianteInformacion2 := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"datos_identificacion?query=Activo:true,TerceroId:"+fmt.Sprintf("%v", idEstudiante)+"&sortby=Id&order=desc&limit=1", &EstudianteInformacion2) //,TipoDocumentoId.Id:3
				if errEstudianteInformacion2 == nil && fmt.Sprintf("%v", EstudianteInformacion2[0]) != "map[]" {
					errProyecto := request.GetJson("http://"+beego.AppConfig.String("ProyectoAcademicoService")+"proyecto_academico_institucion?query=Activo:true,Id:"+fmt.Sprintf("%v", id_proyecto), &proyecto)
					if errProyecto == nil && fmt.Sprintf("%v", proyecto[0]) != "map[]" {
						errCalendario := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario?query=Activo:true,Nivel:"+fmt.Sprintf("%v", proyecto[0].(map[string]interface{})["NivelFormacionId"].(map[string]interface{})["Id"]), &calendario)
						if errCalendario == nil && fmt.Sprintf("%v", calendario[0]) != "map[]" {
							errPeriodo := request.GetJson("http://"+beego.AppConfig.String("ParametroService")+"periodo/"+fmt.Sprintf("%v", calendario[0].(map[string]interface{})["PeriodoId"]), &periodo)
							if errPeriodo == nil && fmt.Sprintf("%v", periodo["Status"]) == "200" {
								errEspaciosAcademicos := request.GetJson("http://"+beego.AppConfig.String("EspaciosAcademicosService")+"espacio-academico-estudiantes?query=activo:true,periodo_id:"+fmt.Sprintf("%v", calendario[0].(map[string]interface{})["Id"])+",estudiante_id:"+fmt.Sprintf("%v", idEstudiante)+"&fields=_id&limit=0", &EspaciosAcademicos)
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

										errInfoNota := request.GetJson("http://"+beego.AppConfig.String("CalificacionesService")+"nota?query=activo:true,estudiante_id:"+fmt.Sprintf("%v", idEstudiante)+"&fields=_id&limit=0", &InfoNota)
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
															InfoNotas.Corte1.Informacion = helpers.PassNotaInfV2(Nota.(map[string]interface{}))
														}
														if EspacioAcademico == EspacioAcademicoNota && estadoRegistro == InfoNotas.Corte2.IdEstado {
															InfoNotas.Corte2.Informacion = helpers.PassNotaInfV2(Nota.(map[string]interface{}))
														}
														if EspacioAcademico == EspacioAcademicoNota && estadoRegistro == InfoNotas.Examen.IdEstado {
															InfoNotas.Examen.Informacion = helpers.PassNotaInfV2(Nota.(map[string]interface{}))
														}
														if EspacioAcademico == EspacioAcademicoNota && estadoRegistro == InfoNotas.Habilit.IdEstado {
															InfoNotas.Habilit.Informacion = helpers.PassNotaInfV2(Nota.(map[string]interface{}))
														}
														if EspacioAcademico == EspacioAcademicoNota && estadoRegistro == InfoNotas.Definitiva.IdEstado {
															InfoNotas.Definitiva.Informacion = helpers.PassNotaInfV2(Nota.(map[string]interface{}))
														}
													}
													NotasAsignaturasEstudiante = append(NotasAsignaturasEstudiante, map[string]interface{}{
														"Grupo":      espAca.(map[string]interface{})["espacio_academico_id"].(map[string]interface{})["grupo"],
														"Asignatura": espAca.(map[string]interface{})["espacio_academico_id"].(map[string]interface{})["nombre"],
														"Creditos":   espAca.(map[string]interface{})["espacio_academico_id"].(map[string]interface{})["creditos"],
														"Corte1":     InfoNotas.Corte1.Informacion,
														"Corte2":     InfoNotas.Corte2.Informacion,
														"Examen":     InfoNotas.Examen.Informacion,
														"Habilit":    InfoNotas.Habilit.Informacion,
														"Definitiva": InfoNotas.Definitiva.Informacion,
														"Acumulado":  helpers.CalculoAcumuladoNotas(InfoNotas),
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

												return resultados, nil

											} else {
												return nil, errors.New("notok2 Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
											}
										} else {
											logs.Error(errInfoNota)
											return nil, errors.New("errInfoNota Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
										}
									} else {
										return nil, errors.New("notok1 Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
									}
								} else {
									logs.Error(errEspaciosAcademicos)
									return nil, errors.New("errEspaciosAcademicos Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
								}
							} else {
								logs.Error(errPeriodo)
								return nil, errors.New("errPeriodo Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
							}
						} else {
							logs.Error(errCalendario)
							return nil, errors.New("errCalendario Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
						}
					} else {
						logs.Error(errProyecto)
						return nil, errors.New("errProyecto Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
					}
				} else {
					logs.Error(errEstudianteInformacion2)
					return nil, errors.New("errEstudianteInformacion2 Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
				}
			} else {
				logs.Error(errEstudianteInformacion1)
				return nil, errors.New("errEstudianteInformacion1 Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
			}
		} else {
			return nil, errors.New("estadosRegistroIDs Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
		}
	} else {
		return nil, errors.New("estadosRegistroIDs Error service GetDatosEstudianteNotas: The request contains an incorrect parameter or no record exist")
	}
}
