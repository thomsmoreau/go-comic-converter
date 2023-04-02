package converter

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

type converterOrder struct {
	name      string
	is_string bool

	section    string
	is_section bool
}

type Converter struct {
	Options *Options
	Cmd     *flag.FlagSet

	order           []converterOrder
	isZeroValueErrs []error
}

func New() *Converter {
	options := NewOptions()
	cmd := flag.NewFlagSet("go-comic-converter", flag.ExitOnError)
	conv := &Converter{
		Options: options,
		Cmd:     cmd,
		order:   make([]converterOrder, 0),
	}

	cmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", filepath.Base(os.Args[0]))
		for _, o := range conv.order {
			if o.is_section {
				fmt.Fprintf(os.Stderr, "\n%s:\n", o.section)
			} else {
				fmt.Fprintln(os.Stderr, conv.Usage(o.is_string, cmd.Lookup(o.name)))
			}
		}
	}

	return conv
}

func (c *Converter) LoadConfig() error {
	return c.Options.LoadDefault()
}

func (c *Converter) AddSection(section string) {
	c.order = append(c.order, converterOrder{section: section, is_section: true})
}

func (c *Converter) AddStringParam(p *string, name string, value string, usage string) {
	c.Cmd.StringVar(p, name, value, usage)
	c.order = append(c.order, converterOrder{name: name, is_string: true})
}

func (c *Converter) AddIntParam(p *int, name string, value int, usage string) {
	c.Cmd.IntVar(p, name, value, usage)
	c.order = append(c.order, converterOrder{name: name})
}

func (c *Converter) AddBoolParam(p *bool, name string, value bool, usage string) {
	c.Cmd.BoolVar(p, name, value, usage)
	c.order = append(c.order, converterOrder{name: name})
}

func (c *Converter) InitParse() {
	c.AddSection("Output")
	c.AddStringParam(&c.Options.Input, "input", "", "Source of comic to convert: directory, cbz, zip, cbr, rar, pdf")
	c.AddStringParam(&c.Options.Output, "output", "", "Output of the epub (directory or epub): (default [INPUT].epub)")
	c.AddStringParam(&c.Options.Author, "author", "GO Comic Converter", "Author of the epub")
	c.AddStringParam(&c.Options.Title, "title", "", "Title of the epub")
	c.AddIntParam(&c.Options.Workers, "workers", runtime.NumCPU(), "Number of workers")
	c.AddBoolParam(&c.Options.Dry, "dry", false, "Dry run to show all options")

	c.AddSection("Config")
	c.AddStringParam(&c.Options.Profile, "profile", c.Options.Profile, fmt.Sprintf("Profile to use: \n%s", c.Options.profiles))
	c.AddIntParam(&c.Options.Quality, "quality", c.Options.Quality, "Quality of the image")
	c.AddBoolParam(&c.Options.Crop, "crop", c.Options.Crop, "Crop images")
	c.AddIntParam(&c.Options.Brightness, "brightness", c.Options.Brightness, "Brightness readjustement: between -100 and 100, > 0 lighter, < 0 darker")
	c.AddIntParam(&c.Options.Contrast, "contrast", c.Options.Contrast, "Contrast readjustement: between -100 and 100, > 0 more contrast, < 0 less contrast")
	c.AddBoolParam(&c.Options.AutoRotate, "autorotate", c.Options.AutoRotate, "Auto Rotate page when width > height")
	c.AddBoolParam(&c.Options.Auto, "auto", false, "Activate all automatic options")
	c.AddBoolParam(&c.Options.AutoSplitDoublePage, "autosplitdoublepage", c.Options.AutoSplitDoublePage, "Auto Split double page when width > height")
	c.AddBoolParam(&c.Options.NoBlankPage, "noblankpage", c.Options.NoBlankPage, "Remove blank pages")
	c.AddBoolParam(&c.Options.Manga, "manga", c.Options.Manga, "Manga mode (right to left)")
	c.AddBoolParam(&c.Options.HasCover, "hascover", c.Options.HasCover, "Has cover. Indicate if your comic have a cover. The first page will be used as a cover and include after the title.")
	c.AddBoolParam(&c.Options.AddPanelView, "addpanelview", c.Options.AddPanelView, "Add an embeded panel view. On kindle you may not need this option as it is handled by the kindle.")
	c.AddIntParam(&c.Options.LimitMb, "limitmb", c.Options.LimitMb, "Limit size of the ePub: Default nolimit (0), Minimum 20")

	c.AddSection("Default config")
	c.AddBoolParam(&c.Options.Show, "show", false, "Show your default parameters")
	c.AddBoolParam(&c.Options.Save, "save", false, "Save your parameters as default")

	c.AddSection("Other")
	c.AddBoolParam(&c.Options.Help, "help", false, "Show this help message")
}

func (c *Converter) Usage(isString bool, f *flag.Flag) string {
	var b strings.Builder
	fmt.Fprintf(&b, "  -%s", f.Name) // Two spaces before -; see next two comments.
	name, usage := flag.UnquoteUsage(f)
	if len(name) > 0 {
		b.WriteString(" ")
		b.WriteString(name)
	}
	// Print the default value only if it differs to the zero value
	// for this flag type.
	if isZero, err := c.isZeroValue(f, f.DefValue); err != nil {
		c.isZeroValueErrs = append(c.isZeroValueErrs, err)
	} else if !isZero {
		if isString {
			fmt.Fprintf(&b, " (default %q)", f.DefValue)
		} else {
			fmt.Fprintf(&b, " (default %v)", f.DefValue)
		}
	}

	// Boolean flags of one ASCII letter are so common we
	// treat them specially, putting their usage on the same line.
	if b.Len() <= 4 { // space, space, '-', 'x'.
		b.WriteString("\t")
	} else {
		// Four spaces before the tab triggers good alignment
		// for both 4- and 8-space tab stops.
		b.WriteString("\n    \t")
	}
	b.WriteString(strings.ReplaceAll(usage, "\n", "\n    \t"))

	return b.String()
}

// isZeroValue determines whether the string represents the zero
// value for a flag.
func (c *Converter) isZeroValue(f *flag.Flag, value string) (ok bool, err error) {
	// Build a zero value of the flag's Value type, and see if the
	// result of calling its String method equals the value passed in.
	// This works unless the Value type is itself an interface type.
	typ := reflect.TypeOf(f.Value)
	var z reflect.Value
	if typ.Kind() == reflect.Pointer {
		z = reflect.New(typ.Elem())
	} else {
		z = reflect.Zero(typ)
	}
	// Catch panics calling the String method, which shouldn't prevent the
	// usage message from being printed, but that we should report to the
	// user so that they know to fix their code.
	defer func() {
		if e := recover(); e != nil {
			if typ.Kind() == reflect.Pointer {
				typ = typ.Elem()
			}
			err = fmt.Errorf("panic calling String method on zero %v for flag %s: %v", typ, f.Name, e)
		}
	}()
	return value == z.Interface().(flag.Value).String(), nil
}

func (c *Converter) Parse() {
	c.Cmd.Parse(os.Args[1:])
	if c.Options.Help {
		c.Cmd.Usage()
		os.Exit(0)
	}

	if c.Options.Auto {
		c.Options.AutoRotate = true
		c.Options.AutoSplitDoublePage = true
	}
}

func (c *Converter) Validate() error {
	// Check input
	if c.Options.Input == "" {
		return errors.New("missing input")
	}

	fi, err := os.Stat(c.Options.Input)
	if err != nil {
		return err
	}

	// Check Output
	var defaultOutput string
	inputBase := filepath.Clean(c.Options.Input)
	if fi.IsDir() {
		defaultOutput = fmt.Sprintf("%s.epub", inputBase)
	} else {
		ext := filepath.Ext(inputBase)
		defaultOutput = fmt.Sprintf("%s.epub", inputBase[0:len(inputBase)-len(ext)])
	}

	if c.Options.Output == "" {
		c.Options.Output = defaultOutput
	}

	c.Options.Output = filepath.Clean(c.Options.Output)
	if filepath.Ext(c.Options.Output) == ".epub" {
		fo, err := os.Stat(filepath.Dir(c.Options.Output))
		if err != nil {
			return err
		}
		if !fo.IsDir() {
			return errors.New("parent of the output is not a directory")
		}
	} else {
		fo, err := os.Stat(c.Options.Output)
		if err != nil {
			return err
		}
		if !fo.IsDir() {
			return errors.New("output must be an existing dir or end with .epub")
		}
		c.Options.Output = filepath.Join(
			c.Options.Output,
			filepath.Base(defaultOutput),
		)
	}

	// Title
	if c.Options.Title == "" {
		ext := filepath.Ext(defaultOutput)
		c.Options.Title = filepath.Base(defaultOutput[0 : len(defaultOutput)-len(ext)])
	}

	// Profile
	if c.Options.Profile == "" {
		return errors.New("profile missing")
	}

	if _, ok := c.Options.profiles[c.Options.Profile]; !ok {
		return fmt.Errorf("profile %q doesn't exists", c.Options.Profile)
	}

	// LimitMb
	if c.Options.LimitMb < 20 && c.Options.LimitMb != 0 {
		return errors.New("limitmb should be 0 or >= 20")
	}

	// Brightness
	if c.Options.Brightness < -100 || c.Options.Brightness > 100 {
		return errors.New("brightness should be between -100 and 100")
	}

	// Contrast
	if c.Options.Contrast < -100 || c.Options.Contrast > 100 {
		return errors.New("contrast should be between -100 and 100")
	}

	return nil
}

func (c Converter) Fatal(err error) {
	c.Cmd.Usage()
	fmt.Fprintf(os.Stderr, "\nError: %s\n", err)
	os.Exit(1)
}