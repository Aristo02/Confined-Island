package screen

import (
	"confinedisland/generator/island"
	"confinedisland/player"
	"confinedisland/sprite"
	"image/color"
)

type Screen struct {
	X, Y           int
	Width, Height  int
	ViewportTilesX int
	ViewportTilesY int
	Background     [][]sprite.Sprite
	player         *player.Player
	world          *island.Island
	lastX, lastY   int
	ZoomLevel      float64
	MinViewport    int
	MaxViewport    int
	Dirty          bool
}

func NewScreen(height int, width int, player *player.Player, world *island.Island) *Screen {
	viewportTilesX := 15
	viewportTilesY := 15
	background := make([][]sprite.Sprite, viewportTilesY)
	for i := range background {
		background[i] = make([]sprite.Sprite, viewportTilesX)
	}
	s := Screen{
		X: 0, Y: 0,
		Height: height, Width: width,
		ViewportTilesX: viewportTilesX,
		ViewportTilesY: viewportTilesY,
		Background:     background,
		player:         player,
		world:          world,
		ZoomLevel:      1.0,
		MinViewport:    7,
		MaxViewport:    31,
		Dirty:          true,
	}
	return &s
}

func (s *Screen) Update() {
	if !s.Dirty {
		return
	}

	// Cache these calculations
	tileWidth := float64(s.Width) / float64(s.ViewportTilesX)
	tileHeight := float64(s.Height) / float64(s.ViewportTilesY)

	cord_y := s.player.WorldY - s.ViewportTilesY/2
	cord_x := s.player.WorldX - s.ViewportTilesX/2

	y_map, x_map := cord_y, cord_x

	for i := 0; i < s.ViewportTilesY; i++ {
		for j := 0; j < s.ViewportTilesX; j++ {
			if y_map >= 0 && y_map < s.world.Height && x_map >= 0 && x_map < s.world.Width {
				s.Background[i][j] = sprite.Sprite{
					Color:  s.world.Background[y_map][x_map],
					X:      float64(j) * tileWidth,
					Y:      float64(i) * tileHeight,
					Width:  tileWidth,
					Height: tileHeight,
				}
			} else {
				s.Background[i][j] = sprite.Sprite{
					Color:  color.RGBA{R: 0, B: 0, G: 0, A: 250},
					X:      float64(j) * tileWidth,
					Y:      float64(i) * tileHeight,
					Width:  tileWidth,
					Height: tileHeight,
				}
			}
			x_map++
		}
		y_map++
		x_map = cord_x
	}

	s.Dirty = false // Reset the flag after update
}

func (s *Screen) GetPlayerPosition() (int, int) {
	return s.player.WorldX, s.player.WorldY
}
