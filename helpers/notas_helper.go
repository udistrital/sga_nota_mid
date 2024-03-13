package helpers

import (
	"fmt"
	"math"
	"reflect"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/udistrital/utils_oas/request"
)

// TipoEstado
type TipoEstado struct {
	IdEstado         string
	Nombre           string
	Existe           bool
	IdRegistroNota   string
	Finalizado       bool
	EditExtemporaneo bool
	Informacion      map[string]interface{}
}

// EstadosRegistro
type EstadosRegistro struct {
	Corte1     TipoEstado
	Corte2     TipoEstado
	Examen     TipoEstado
	Habilit    TipoEstado
	Definitiva TipoEstado
}

func ObtenerEspaciosAcademicos(idDocente string) (map[string]interface{}, error) {
	var registros map[string]interface{}
	url := fmt.Sprintf("http://%sespacio-academico?query=activo:true,docente_id:%v", beego.AppConfig.String("EspaciosAcademicosService"), idDocente)
	fmt.Println(url)
	err := request.GetJson(url, &registros)
	return registros, err
}

func ObtenerProyectos() ([]interface{}, error) {
	var proyectos []interface{}
	url := fmt.Sprintf("http://%s/proyecto_academico_institucion?query=Activo:true&fields=Id,Nombre,NivelFormacionId&limit=0", beego.AppConfig.String("ProyectoAcademicoService"))
	err := request.GetJson(url, &proyectos)
	return proyectos, err
}

func ObtenerCalendarios() ([]interface{}, error) {
	var calendarios []interface{}
	url := fmt.Sprintf("http://%s/calendario?query=Activo:true&fields=Id,Nombre,PeriodoId&limit=0", beego.AppConfig.String("EventoService"))
	fmt.Println(url)
	err := request.GetJson(url, &calendarios)
	return calendarios, err
}

func ObtenerPeriodos() (map[string]interface{}, error) {
	var periodos map[string]interface{}
	url := fmt.Sprintf("http://%s/periodo?query=Activo:true&fields=Id,Nombre&limit=0", beego.AppConfig.String("ParametroService"))
	err := request.GetJson(url, &periodos)
	return periodos, err
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

func GetProyectoFromEspacioAcademico_temporal(id_estudiante string) (string, bool) {
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

func CalculoNotasPorCortes(Notas map[string]interface{}) map[string]interface{} {

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

func CalculoDefinitiva(NotasEst interface{}) float64 {

	var calculo = 0.0

	n1 := NotasEst.(map[string]interface{})["Corte1"].(map[string]interface{})["data"].(map[string]interface{})["nota_definitiva"].(float64)
	n2 := NotasEst.(map[string]interface{})["Corte2"].(map[string]interface{})["data"].(map[string]interface{})["nota_definitiva"].(float64)
	n3 := NotasEst.(map[string]interface{})["Examen"].(map[string]interface{})["data"].(map[string]interface{})["nota_definitiva"].(float64)
	n4 := NotasEst.(map[string]interface{})["Habilit"].(map[string]interface{})["data"].(map[string]interface{})["nota_definitiva"].(float64)

	calculo = n1 + n2 + n3 + n4

	Definitiva := math.Round(calculo*10) / 10

	return Definitiva
}

func CalculoAcumuladoNotas(calif EstadosRegistro) float64 {

	calculo := 0.0

	n1 := calif.Corte1.Informacion["data"].(map[string]interface{})["nota_definitiva"].(float64)
	n2 := calif.Corte2.Informacion["data"].(map[string]interface{})["nota_definitiva"].(float64)
	n3 := calif.Examen.Informacion["data"].(map[string]interface{})["nota_definitiva"].(float64)
	n4 := calif.Habilit.Informacion["data"].(map[string]interface{})["nota_definitiva"].(float64)

	calculo = n1 + n2 + n3 + n4

	Acumulado := math.Floor(calculo*100) / 100

	return Acumulado
}

func PassNotaInf(N map[string]interface{}) map[string]interface{} {
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

func PassNotaInfV2(N map[string]interface{}) map[string]interface{} {
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

func PassNotaEmpty() map[string]interface{} {
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

func PassPorcentajeEmpty(reg string) map[string]interface{} {
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

func FindNamebyId(list []interface{}, id string) string {
	for _, item := range list {
		if fmt.Sprintf("%v", item.(map[string]interface{})["Id"]) == id {
			return fmt.Sprintf("%v", item.(map[string]interface{})["Nombre"])
		}
	}
	return ""
}

func FindIdsbyId(list []interface{}, id string) map[string]interface{} {
	for _, item := range list {
		if fmt.Sprintf("%v", item.(map[string]interface{})["Id"]) == id {
			return item.(map[string]interface{})
		}
	}
	return map[string]interface{}{}
}

func ValidatePutPorcentajes(p map[string]interface{}) bool {
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

func ValidatePutNotasEstudiantes(n map[string]interface{}) bool {
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

func VerificarNil(v interface{}) bool {
	// Un valor nil en Go tiene un tipo inválido y un valor de puntero nulo.
	if v == nil {
		return true
	}

	// Usar reflect para ver si el valor es realmente nil (útil para interfaces y punteros).
	valor := reflect.ValueOf(v)
	if valor.Kind() == reflect.Ptr || valor.Kind() == reflect.Interface {
		if valor.IsNil() {
			return true
		}
	}

	return false
}
