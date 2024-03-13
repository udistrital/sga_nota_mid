package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

    beego.GlobalControllerRouter["github.com/udistrital/sga_nota_mid/controllers:NotasController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_nota_mid/controllers:NotasController"],
        beego.ControllerComments{
            Method: "GetDatosDocenteAsignatura",
            Router: "/asignaturas/:id_asignatura/info-docente",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_nota_mid/controllers:NotasController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_nota_mid/controllers:NotasController"],
        beego.ControllerComments{
            Method: "GetModificacionExtemporanea",
            Router: "/asignaturas/:id_asignatura/modificacion-extemporanea",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_nota_mid/controllers:NotasController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_nota_mid/controllers:NotasController"],
        beego.ControllerComments{
            Method: "GetPorcentajesAsignatura",
            Router: "/asignaturas/:id_asignatura/periodos/:id_periodo/porcentajes",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_nota_mid/controllers:NotasController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_nota_mid/controllers:NotasController"],
        beego.ControllerComments{
            Method: "PutPorcentajesAsignatura",
            Router: "/asignaturas/porcentajes",
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_nota_mid/controllers:NotasController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_nota_mid/controllers:NotasController"],
        beego.ControllerComments{
            Method: "GetEspaciosAcademicosDocente",
            Router: "/docentes/:id_docente/espacios-academicos",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_nota_mid/controllers:NotasController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_nota_mid/controllers:NotasController"],
        beego.ControllerComments{
            Method: "GetCapturaNotas",
            Router: "/notas/asignaturas/:id_asignatura/periodos/:id_periodo/estudiantes",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_nota_mid/controllers:NotasController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_nota_mid/controllers:NotasController"],
        beego.ControllerComments{
            Method: "PutCapturaNotas",
            Router: "/notas/asignaturas/estudiantes",
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_nota_mid/controllers:NotasController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_nota_mid/controllers:NotasController"],
        beego.ControllerComments{
            Method: "GetDatosEstudianteNotas",
            Router: "/notas/estudiantes/:id_estudiante",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_nota_mid/controllers:NotasController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_nota_mid/controllers:NotasController"],
        beego.ControllerComments{
            Method: "GetEstadosRegistros",
            Router: "/periodos/:id_periodo/estados-registros",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
