package contacts

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/xuri/excelize/v2"
)

const hojaPlantilla = "Contactos"

var ErrExcelSinFilas = errors.New("el archivo de Excel no tiene filas")

var columnasPlantilla = []string{"nombre", "empresa", "correo", "pais", "cargo"}

var ejemploPlantilla = [][]string{
	{"Ana Quispe", "Agroindustrias del Sur SAC", "ana.quispe@ejemplo.pe", "Perú", "Gerente de Compras"},
	{"Luis Mamani", "Transportes Interoceánica EIRL", "luis.mamani@ejemplo.pe", "Perú", "Jefe de Logística"},
}

func LeerExcel(lector io.Reader) ([][]string, error) {
	libro, err := excelize.OpenReader(lector)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el archivo de Excel: %w", err)
	}
	defer libro.Close()

	hojas := libro.GetSheetList()
	if len(hojas) == 0 {
		return nil, ErrExcelSinFilas
	}

	filas, err := libro.GetRows(hojas[0])
	if err != nil {
		return nil, fmt.Errorf("no se pudieron leer las filas: %w", err)
	}

	if len(filas) == 0 {
		return nil, ErrExcelSinFilas
	}

	return filas, nil
}

func GenerarPlantilla() (*excelize.File, error) {
	libro := excelize.NewFile()

	indice, err := libro.NewSheet(hojaPlantilla)
	if err != nil {
		return nil, fmt.Errorf("no se pudo crear la hoja: %w", err)
	}

	libro.SetActiveSheet(indice)
	if err := libro.DeleteSheet("Sheet1"); err != nil {
		return nil, fmt.Errorf("no se pudo preparar la plantilla: %w", err)
	}

	estiloCabecera, err := libro.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"18181B"}, Pattern: 1},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("no se pudo aplicar el estilo: %w", err)
	}

	for posicion, columna := range columnasPlantilla {
		celda, err := excelize.CoordinatesToCellName(posicion+1, 1)
		if err != nil {
			return nil, fmt.Errorf("no se pudo ubicar la celda: %w", err)
		}

		if err := libro.SetCellStr(hojaPlantilla, celda, columna); err != nil {
			return nil, fmt.Errorf("no se pudo escribir la cabecera: %w", err)
		}

		letra, err := excelize.ColumnNumberToName(posicion + 1)
		if err != nil {
			return nil, fmt.Errorf("no se pudo ubicar la columna: %w", err)
		}

		if err := libro.SetColWidth(hojaPlantilla, letra, letra, anchoColumna(columna)); err != nil {
			return nil, fmt.Errorf("no se pudo ajustar el ancho: %w", err)
		}
	}

	if err := libro.SetCellStyle(hojaPlantilla, "A1", ultimaCeldaCabecera(), estiloCabecera); err != nil {
		return nil, fmt.Errorf("no se pudo aplicar el estilo: %w", err)
	}

	for numeroFila, ejemplo := range ejemploPlantilla {
		for posicion, contenido := range ejemplo {
			celda, err := excelize.CoordinatesToCellName(posicion+1, numeroFila+2)
			if err != nil {
				return nil, fmt.Errorf("no se pudo ubicar la celda: %w", err)
			}

			if err := libro.SetCellStr(hojaPlantilla, celda, contenido); err != nil {
				return nil, fmt.Errorf("no se pudo escribir el ejemplo: %w", err)
			}
		}
	}

	if err := libro.SetPanes(hojaPlantilla, &excelize.Panes{
		Freeze:      true,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	}); err != nil {
		return nil, fmt.Errorf("no se pudo fijar la cabecera: %w", err)
	}

	return libro, nil
}

func anchoColumna(columna string) float64 {
	switch columna {
	case "empresa":
		return 34
	case "correo":
		return 30
	case "nombre", "cargo":
		return 24
	default:
		return 14
	}
}

func ultimaCeldaCabecera() string {
	letra, err := excelize.ColumnNumberToName(len(columnasPlantilla))
	if err != nil {
		return "A1"
	}
	return fmt.Sprintf("%s1", letra)
}

func esArchivoExcel(nombreArchivo string) bool {
	return strings.HasSuffix(strings.ToLower(nombreArchivo), ".xlsx")
}
