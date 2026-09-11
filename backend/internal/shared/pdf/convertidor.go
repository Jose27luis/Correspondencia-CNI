package pdf

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

const tiempoMaximoConversion = 90 * time.Second

var ErrNoDisponible = errors.New("el conversor a PDF no está instalado en el servidor")

type Convertidor struct {
	ruta  string
	mutex sync.Mutex
}

func NuevoConvertidor(ruta string) *Convertidor {
	if _, err := os.Stat(ruta); err != nil {
		return &Convertidor{}
	}
	return &Convertidor{ruta: ruta}
}

func (c *Convertidor) Disponible() bool {
	return c.ruta != ""
}

func (c *Convertidor) ConvertirDOCX(ctx context.Context, documento []byte) ([]byte, error) {
	if !c.Disponible() {
		return nil, ErrNoDisponible
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	directorio, err := os.MkdirTemp("", "correspondencia-pdf-")
	if err != nil {
		return nil, fmt.Errorf("no se pudo preparar la conversión: %w", err)
	}
	defer os.RemoveAll(directorio)

	entrada := filepath.Join(directorio, "carta.docx")
	if err := os.WriteFile(entrada, documento, 0o600); err != nil {
		return nil, fmt.Errorf("no se pudo preparar la conversión: %w", err)
	}

	contexto, cancelar := context.WithTimeout(ctx, tiempoMaximoConversion)
	defer cancelar()

	comando := exec.CommandContext(
		contexto,
		c.ruta,
		"-env:UserInstallation=file://"+filepath.Join(directorio, "perfil"),
		"--headless",
		"--norestore",
		"--nolockcheck",
		"--convert-to", "pdf",
		"--outdir", directorio,
		entrada,
	)
	comando.Env = append(os.Environ(), "HOME="+directorio)

	if salida, err := comando.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("la conversión a PDF falló: %w: %s", err, string(salida))
	}

	resultado, err := os.ReadFile(filepath.Join(directorio, "carta.pdf"))
	if err != nil {
		return nil, fmt.Errorf("la conversión no produjo el PDF: %w", err)
	}

	return resultado, nil
}
