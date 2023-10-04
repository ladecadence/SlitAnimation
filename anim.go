package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"os"

	"github.com/pwiecz/go-fltk"
)

// IMAGE FUNCTIONS

func imageSize(img image.Image) (int, int) {
	if img == nil {
		return 0, 0
	}
	return img.Bounds().Max.X, img.Bounds().Max.Y
}

func openImage(path string) (image.Image, error) {
	// open image
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer f.Close()

	// decode it
	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Println("Decoding error:", err.Error())
		return nil, err
	}

	return img, nil
}

func generate_anim(files []string, path string, bar_width int) error {
	// image list
	images := []image.Image{}

	// open first image
	img, err := openImage(files[0])
	if err != nil {
		fmt.Printf("Error opening image %s\n", files[0])
		return errors.New("Error opening image")
		//os.Exit(1)
	}
	images = append(images, img)

	// save size
	width, height := imageSize(images[0])
	fmt.Printf("Image: w:%d, h:%d\n", width, height)

	// check bars
	if width%bar_width != 0 {
		fmt.Println("Bar width must be a divider of images' width")
		return errors.New("Bar width must be a divider of images' width")
		//os.Exit(1)
	}

	for i := 1; i < len(files); i++ {
		img, err := openImage(files[i])
		if err != nil {
			fmt.Printf("Error opening image %s\n", files[i])
			return errors.New("Error opening image")
			//os.Exit(1)
		}
		if img.Bounds().Max.X != width || img.Bounds().Max.Y != height {
			fmt.Printf("Images must be the same size")
			return errors.New("Images must be of the same size")
			//os.Exit(1)
		}
		images = append(images, img)
	}

	// ok, create new image and loop values
	fmt.Printf("Number of images: %d\n", len(images))
	out_img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{width, height}})
	// initialize to pink to see problems
	magenta := color.RGBA{255, 0, 255, 255}
	draw.Draw(out_img, out_img.Bounds(), &image.Uniform{magenta}, image.ZP, draw.Src)

	img_count := len(images)
	bar_count := width / bar_width
	fmt.Printf("Number of bars: %d\n", bar_count)

	// loop
	for barn := 0; barn < bar_count; barn++ {
		dest_min_point := image.Point{barn * bar_width, 0}
		dest_max_point := image.Point{dest_min_point.X + bar_width, height}
		rect := image.Rectangle{dest_min_point, dest_max_point}
		draw.Draw(out_img, rect, images[barn%img_count], dest_min_point, draw.Src)
	}

	// write image
	f, err := os.Create(path + "output.png")
	if err != nil {
		fmt.Println("Error creating output image")
		return errors.New("Error creating output image")
	}
	defer f.Close()
	png.Encode(f, out_img)

	// create mask
	mask_img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{width * 1, height}})
	black := color.RGBA{0, 0, 0, 255}
	white := color.RGBA{255, 255, 255, 255}

	draw.Draw(mask_img, mask_img.Bounds(), &image.Uniform{black}, image.ZP, draw.Src)

	for i := 0; i < (bar_count/bar_width)*1; i++ {
		dest_min_point := image.Point{img_count * bar_width * i, 0}
		dest_max_point := image.Point{dest_min_point.X + bar_width, height}
		rect := image.Rectangle{dest_min_point, dest_max_point}
		draw.Draw(mask_img, rect, &image.Uniform{white}, image.ZP, draw.Src)
	}

	// write mask
	fm, err := os.Create(path + "mask.png")
	if err != nil {
		fmt.Println("Error creating output image")
		return errors.New("Error creating output image")
	}
	defer fm.Close()
	png.Encode(fm, mask_img)

	return nil
}

func main() {
	// command line args
	// if len(os.Args) < 4 {
	// 	fmt.Println("Usage:")
	// 	fmt.Printf("%s <bar width> <img 1> <img 2> ...\n", os.Args[0])
	// 	fmt.Println("\tWhere <bar width> is in pixels")
	// 	os.Exit(1)
	// }

	// bar_width, err := strconv.Atoi(os.Args[1])
	// if err != nil {
	// 	fmt.Println("bar width must be a number")
	// 	os.Exit(1)
	// }
	// fmt.Printf("Bar width: %d\n", bar_width)

	// generate_anim(os.Args[2:], "out/", bar_width)

	list_imgs := []string{}

	// window
	//fltk.SetScheme("gleam")
	win := fltk.NewWindow(400, 360)
	win.SetLabel("Animasiones")

	// image list
	list_label := fltk.NewBox(fltk.NO_BOX, 5, 5, 390, 20, "Images: ")
	list_label.SetLabelFont(fltk.ITALIC)
	list := fltk.NewTextDisplay(5, 30, 390, 250, "")
	list_text := fltk.NewTextBuffer()
	list.SetBuffer(list_text)

	// list buttons
	button_add := fltk.NewButton(5, 285, 185, 20, "Add images")
	button_add.SetCallback(func() {
		nfc := fltk.NewNativeFileChooser()
		defer nfc.Destroy()
		nfc.SetOptions(fltk.NativeFileChooser_PREVIEW | fltk.NativeFileChooser_NEW_FOLDER)
		nfc.SetType(fltk.NativeFileChooser_BROWSE_MULTI_FILE)
		nfc.SetDirectory("")
		nfc.SetFilter("Images\t*.{png,jpg}")
		nfc.SetTitle("Select images...")
		nfc.Show()
		for _, filename := range nfc.Filenames() {
			list.InsertText(filename)
			list.InsertText("\n")
		}
		list_imgs = append(list_imgs, nfc.Filenames()...)
		fmt.Printf("%v\n", list_imgs)

	})

	button_clear := fltk.NewButton(210, 285, 185, 20, "Clear")
	button_clear.SetColor(fltk.ColorFromRgb(250, 50, 50))
	button_clear.SetLabelColor(fltk.ColorFromRgb(250, 250, 250))
	button_clear.SetCallback(func() {
		list_imgs = []string{}
		list.Buffer().SetText("")
		fmt.Printf("%v\n", list_imgs)
	})

	// width spinner
	width_spin := fltk.NewSpinner(5, 310, 100, 40, "")
	width_spin.SetType(0)
	width_spin.SetMaximum(50)
	width_spin.SetMinimum(1)

	// Run
	button_run := fltk.NewButton(115, 310, 280, 40, "Create Animation")
	button_run.SetLabelFont(fltk.BOLD)
	button_run.SetColor(fltk.ColorFromRgb(50, 250, 50))
	button_run.SetCallback(func() {
		if len(list_imgs) > 2 {
			nfc := fltk.NewNativeFileChooser()
			defer nfc.Destroy()
			nfc.SetOptions(fltk.NativeFileChooser_PREVIEW | fltk.NativeFileChooser_NEW_FOLDER)
			nfc.SetType(fltk.NativeFileChooser_BROWSE_DIRECTORY)
			nfc.SetDirectory("")
			nfc.SetTitle("Select output folder...")
			nfc.Show()

			if len(nfc.Filenames()) != 0 {
				err := generate_anim(list_imgs, nfc.Filenames()[0]+"/", int(width_spin.Value()))
				if err != nil {
					fltk.MessageBox("Error", err.Error())
				} else {
					fltk.MessageBox("Ok", "Animation generated on "+nfc.Filenames()[0])
				}
			}
		} else {
			fltk.MessageBox("Error", "Need at least two images")
		}
	})

	win.End()
	win.Show()
	fltk.Run()

}
