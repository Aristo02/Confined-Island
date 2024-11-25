package main

import (
	"confinedisland/generator/island"
	"confinedisland/player"
	"confinedisland/screen"
	"confinedisland/sprite"
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type CachedTile struct {
	image    *ebiten.Image
	lastUsed time.Time
}

type TileCache struct {
	tiles   map[color.RGBA]*CachedTile
	maxSize int
}

type Game struct {
	world       *island.Island
	scene       *screen.Screen
	player      *player.Player
	tileCache   *TileCache
	playerImage *ebiten.Image
}

func NewGame(world *island.Island, scene *screen.Screen, player *player.Player) *Game {
	g := &Game{
		world:     world,
		scene:     scene,
		player:    player,
		tileCache: NewTileCache(100),
	}

	g.playerImage = ebiten.NewImage(1, 1)
	g.playerImage.Fill(color.RGBA{R: 250, G: 0, B: 0, A: 50})

	return g
}

func (g *Game) Update() error {
	// Vérification des touches fléchées
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		if g.player.WorldY > 0 {
			g.player.WorldY -= 1
			g.scene.Dirty = true
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		if g.player.WorldY < g.world.Height-1 {
			g.player.WorldY += 1
			g.scene.Dirty = true
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		if g.player.WorldX > 0 {
			g.player.WorldX -= 1
			g.scene.Dirty = true
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		if g.player.WorldX < g.world.Width-1 {
			g.player.WorldX += 1
			g.scene.Dirty = true
		}
	}

	// Add zoom controls
	if ebiten.IsKeyPressed(ebiten.KeyO) { // Zoom in with 'o'
		g.scene.ZoomLevel *= 1.02 // Increase by 2%
		g.scene.Dirty = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyP) { // Zoom out with 'p'
		g.scene.ZoomLevel *= 0.98 // Decrease by 2%
		g.scene.Dirty = true
	}

	// Add viewport size controls
	if ebiten.IsKeyPressed(ebiten.KeyK) { // Decrease viewport
		if g.scene.ViewportTilesX > g.scene.MinViewport {
			newSize := g.scene.ViewportTilesX - 1
			g.resizeViewport(newSize)
			g.scene.Dirty = true
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyL) { // Increase viewport
		if g.scene.ViewportTilesX < g.scene.MaxViewport {
			newSize := g.scene.ViewportTilesX + 1
			g.resizeViewport(newSize)
			g.scene.Dirty = true
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		playerX, playerY := g.scene.GetPlayerPosition()

		// Get the world coordinates for each tile in the center 3x3 of viewport
		centerX := g.scene.ViewportTilesX / 2
		centerY := g.scene.ViewportTilesY / 2

		fmt.Printf("\nViewport size: %dx%d\n", g.scene.ViewportTilesX, g.scene.ViewportTilesY)
		fmt.Printf("Player world pos: (%d,%d)\n", playerX, playerY)

		// Show the 3x3 grid of world coordinates around where player should be
		for y := -1; y <= 1; y++ {
			for x := -1; x <= 1; x++ {
				viewportTile := fmt.Sprintf("(%d,%d)", centerX+x, centerY+y)
				worldX := playerX + x
				worldY := playerY + y
				worldTile := fmt.Sprintf("(%d,%d)", worldX, worldY)
				if x == 0 && y == 0 {
					fmt.Printf("* %s -> %s *\n", viewportTile, worldTile)
				} else {
					fmt.Printf("%s -> %s\n", viewportTile, worldTile)
				}
			}
		}
	}

	g.scene.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 255})

	baseTileSize := math.Min(
		float64(g.scene.Width)/float64(g.scene.ViewportTilesX),
		float64(g.scene.Height)/float64(g.scene.ViewportTilesY),
	)
	tileSize := math.Floor(baseTileSize * g.scene.ZoomLevel)

	offsetX := math.Floor((float64(g.scene.Width) - (tileSize * float64(g.scene.ViewportTilesX))) / 2)
	offsetY := math.Floor((float64(g.scene.Height) - (tileSize * float64(g.scene.ViewportTilesY))) / 2)

	// Draw background tiles using cached images
	op := &ebiten.DrawImageOptions{}
	for i, row := range g.scene.Background {
		for j, tile := range row {
			tileImg := g.tileCache.Get(tile.Color)
			if tileImg == nil {
				// Create new image if it doesn't exist in cache
				tileImg = ebiten.NewImage(1, 1)
				tileImg.Fill(tile.Color)
				g.tileCache.Put(tile.Color, tileImg)
			}

			op.GeoM.Reset()
			op.GeoM.Scale(tileSize, tileSize)
			op.GeoM.Translate(offsetX+float64(j)*tileSize, offsetY+float64(i)*tileSize)
			screen.DrawImage(tileImg, op)
		}
	}

	// Draw player
	if g.playerImage == nil {
		g.playerImage = ebiten.NewImage(1, 1)
		g.playerImage.Fill(color.RGBA{R: 250, G: 0, B: 0, A: 50})
	}

	op.GeoM.Reset()
	op.GeoM.Scale(tileSize, tileSize)
	centerTileX := g.scene.ViewportTilesX / 2
	centerTileY := g.scene.ViewportTilesY / 2
	op.GeoM.Translate(
		offsetX+float64(centerTileX)*tileSize,
		offsetY+float64(centerTileY)*tileSize,
	)
	screen.DrawImage(g.playerImage, op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func (g *Game) resizeViewport(newSize int) {
	if newSize == g.scene.ViewportTilesX {
		return
	}

	// Create new background with correct size
	newBackground := make([][]sprite.Sprite, newSize)
	for i := range newBackground {
		newBackground[i] = make([]sprite.Sprite, newSize)
	}
	g.scene.Background = newBackground
	g.scene.ViewportTilesX = newSize
	g.scene.ViewportTilesY = newSize
}

func NewTileCache(maxSize int) *TileCache {
	return &TileCache{
		tiles:   make(map[color.RGBA]*CachedTile),
		maxSize: maxSize,
	}
}

func (tc *TileCache) Get(c color.RGBA) *ebiten.Image {
	if cached, exists := tc.tiles[c]; exists {
		cached.lastUsed = time.Now()
		return cached.image
	}
	return nil
}

func (tc *TileCache) Put(c color.RGBA, img *ebiten.Image) {
	// If cache is full, remove least recently used entry
	if len(tc.tiles) >= tc.maxSize {
		var oldestColor color.RGBA
		oldestTime := time.Now()

		for color, tile := range tc.tiles {
			if tile.lastUsed.Before(oldestTime) {
				oldestTime = tile.lastUsed
				oldestColor = color
			}
		}
		delete(tc.tiles, oldestColor)
	}

	tc.tiles[c] = &CachedTile{
		image:    img,
		lastUsed: time.Now(),
	}
}

func main() {
	islandConf := island.IslandConfig{Width: 50, Height: 50}
	width, height := 960, 512
	world := island.NewIsland(islandConf)

	// Calculate center position in tiles
	centerTileX := float64(width) / 2
	centerTileX = float64(int(centerTileX/32) * 32) // Snap to grid
	centerTileY := float64(height) / 2
	centerTileY = float64(int(centerTileY/32) * 32) // Snap to grid

	player := &player.Player{
		X:      centerTileX,
		Y:      centerTileY,
		WorldX: int(islandConf.Width)/2 - 1,
		WorldY: int(islandConf.Height)/2 - 1,
	}

	scene := screen.NewScreen(int(height), int(width), player, world)
	g := NewGame(world, scene, player)

	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowTitle("Mon premier jeu Ebitengine")

	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}

}
