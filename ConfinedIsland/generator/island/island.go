package island

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"time"

	"github.com/aquilax/go-perlin"
	"golang.org/x/exp/rand"
)

type IslandConfig struct {
	Width  int
	Height int
	Seed   int64
}

type Island struct {
	Width      int
	Height     int
	Seed       int64
	Base       [][]float64
	Background [][]color.RGBA
}

func NewIsland(config IslandConfig) *Island {
	// Initialiser le générateur de nombres aléatoires...
	background := make([][]color.RGBA, config.Height)
	if config.Seed == 0 {
		// Initialiser le générateur de nombres aléatoires
		rand.Seed(uint64(time.Now().UnixNano()))
		// Définir la plage
		min := -100000 * 100000 // valeur minimale
		max := 100000 * 100000  // valeur maximale

		// Générer un nombre aléatoire entre min et max
		config.Seed = int64(rand.Intn(max-min+1) + min)
	}
	for i := range background {
		background[i] = make([]color.RGBA, config.Width)
	}
	i := Island{Width: config.Width, Height: config.Height, Seed: config.Seed, Background: background}
	i.Create_island() // Crée l'île
	i.Create_image()  // Remplit l'arrière-plan avec les couleurs
	return &i
}

func (i *Island) Create_island() {

	// Créer un générateur de bruit de Perlin
	p := perlin.NewPerlin(0.5, 0.5, 2, i.Seed)
	i.Base = make([][]float64, i.Height) // Initialiser Base ici

	for x := 0; x < i.Height; x++ {
		ligne := make([]float64, i.Width) // Créez la ligne avec la largeur

		for y := 0; y < i.Width; y++ {
			// Calculer la distance du pixel par rapport au centre
			distFromCenter := math.Sqrt(float64((x-i.Width/2)*(x-i.Width/2)+(y-i.Height/2)*(y-i.Height/2))) / (float64(i.Width) / 2)

			// Calculer la valeur de bruit de Perlin pour chaque pixel
			value := p.Noise2D(float64(x)/(float64(i.Width)/10), float64(y)/(float64(i.Height)/10))

			// Ajuster la valeur de bruit pour créer une île centrale avec différents reliefs
			value = (value + 1.0) / 2.0
			value = value * value * (3 - 2*value)
			value = value * (1.0 - distFromCenter)
			ligne[y] = value
		}
		i.Base[x] = ligne
	}
}

func (i *Island) Create_image() {
	img := image.NewRGBA(image.Rect(0, 0, i.Width, i.Height))

	for x := 0; x < i.Height; x++ {
		for y := 0; y < i.Width; y++ {
			value := i.Base[x][y]
			var r, g, b uint8
			// Définir la couleur du pixel en fonction de la valeur de bruit
			if value < 0.04 {
				r, g, b = 18, 60, 93 // Bleu foncé pour l'eau profonde
			} else if value < 0.07 {
				r, g, b = 43, 136, 194 // Bleu clair pour l'eau peu profonde
			} else if value < 0.10 {
				r, g, b = 250, 250, 200 // Beige pour la plage
			} else if value < 0.15 {
				r, g, b = 250, 200, 100 // Jaune pour le désert
			} else if value < 0.28 {
				r, g, b = 150, 200, 100 // Vert clair pour les plaines
			} else if value < 0.40 {
				r, g, b = 100, 150, 0 // Vert foncé pour la forêt
			} else if value < 0.5 {
				r, g, b = 200, 200, 200 // Gris pour les montagnes
			} else if value < 0.6 {
				r, g, b = 255, 255, 255 // Blanc pour les sommets enneigés
			} else if value < 0.7 {
				r, g, b = 50, 50, 50 // Obsidienne
			} else {
				r, g, b = 217, 42, 0 // Lave
			}
			color := color.RGBA{r, g, b, 255}
			img.Set(x, y, color)
			i.Background[y][x] = color // Remplissez Background correctement
		}
	}
	// Enregistrer l'image dans un fichier...
	// Enregistrer l'image dans un fichier
	file, err := os.Create("island.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		panic(err)
	}
}
