package cobra2vitepress

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const markdownExtension = ".md"

func printOptions(buf *bytes.Buffer, cmd *cobra.Command, name string) error {
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		buf.WriteString("### Options\n\n```\n")
		flags.PrintDefaults()
		buf.WriteString("```\n\n")
	}

	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		buf.WriteString("### Options inherited from parent commands\n\n```\n")
		parentFlags.PrintDefaults()
		buf.WriteString("```\n\n")
	}
	return nil
}

// GenMarkdown creates markdown output.
func GenMarkdown(cmd *cobra.Command, w io.Writer) error {
	return GenMarkdownCustom(cmd, w, func(s string) string { return s })
}

// GenMarkdownCustom creates custom markdown output.
func GenMarkdownCustom(cmd *cobra.Command, w io.Writer, linkHandler func(string) string) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)
	name := cmd.CommandPath()
	k := strings.Split(name, " ")
	if len(k) > 1 {
		name = strings.Join(k[1:], " ")
	}
	buf.WriteString(fmt.Sprintf("---\nsidebarTitle: %s\nsidebarOrder: 1\n---\n\n", name))
	if len(cmd.Version) > 0 {
		buf.WriteString("## " + name + "<Badge type=\"tip\" text=\"" + cmd.Version + "\" />\n\n")
	} else {
		buf.WriteString("## " + name + "\n\n")
	}
	buf.WriteString(cmd.Short + "\n\n")
	if len(cmd.Long) > 0 {
		buf.WriteString("### Synopsis\n\n")
		buf.WriteString(cmd.Long + "\n\n")
	}

	if cmd.Runnable() {
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", cmd.UseLine()))
	}

	if len(cmd.Example) > 0 {
		buf.WriteString("### Examples\n\n")
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", cmd.Example))
	}

	if err := printOptions(buf, cmd, name); err != nil {
		return err
	}
	if hasSeeAlso(cmd) {
		buf.WriteString("### SEE ALSO\n\n")
		if cmd.HasParent() {
			parent := cmd.Parent()
			pname := parent.CommandPath()
			link := pname + markdownExtension
			link = strings.ReplaceAll(link, " ", "_")
			if len(strings.Split(link, "_")) > 1 {
				link = strings.ReplaceAll(link, k[0]+"_", "")
			} else {
				link = "../" + link
			}
			buf.WriteString(fmt.Sprintf("* [%s](%s)\t - %s\n", pname, linkHandler(link), parent.Short))
			cmd.VisitParents(func(c *cobra.Command) {
				if c.DisableAutoGenTag {
					cmd.DisableAutoGenTag = c.DisableAutoGenTag
				}
			})
		}

		children := cmd.Commands()
		sort.Sort(byName(children))

		for _, child := range children {
			if !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
				continue
			}
			cname := name + " " + child.Name()
			link := cname + markdownExtension
			link = strings.ReplaceAll(link, " ", "_")
			if len(strings.Split(link, "_")) > 1 {
				link = strings.ReplaceAll(link, k[0]+"_", "")
			}
			buf.WriteString(fmt.Sprintf("* [%s](%s)\t - %s\n", cname, linkHandler(link), child.Short))
		}
		buf.WriteString("\n")
	}
	if !cmd.DisableAutoGenTag {
		buf.WriteString("::: tip\n>Auto generated by spf13/cobra on " + time.Now().Format("2006-Jan-2") + "\n:::")
	}
	_, err := buf.WriteTo(w)
	return err
}

// GenMarkdownTree will generate a markdown page for this command and all
// descendants in the directory given. The header may be nil.
// This function may not work correctly if your command names have `-` in them.
// If you have `cmd` with two subcmds, `sub` and `sub-third`,
// and `sub` has a subcommand called `third`, it is undefined which
// help output will be in the file `cmd-sub-third.1`.
func GenMarkdownTree(cmd *cobra.Command, dir string) error {
	identity := func(s string) string { return s }
	emptyStr := func(s string) string { return "" }
	return GenMarkdownTreeCustom(cmd, dir, emptyStr, identity)
}

// GenMarkdownTreeCustom is the the same as GenMarkdownTree, but
// with custom filePrepender and linkHandler.
func GenMarkdownTreeCustom(cmd *cobra.Command, dir string, filePrepender, linkHandler func(string) string) error {
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := GenMarkdownTreeCustom(c, dir, filePrepender, linkHandler); err != nil {
			return err
		}
	}

	cpath := strings.Split(cmd.CommandPath(), " ")
	basename := cpath[0] + markdownExtension
	filename := filepath.Join(dir, basename)
	if len(cpath) > 1 {
		_ = os.MkdirAll(filepath.Join(dir, cpath[1]), 0755)
		if len(cpath) == 2 {
			basename = cpath[1] + markdownExtension
		} else {
			basename = strings.Join(cpath[1:], "_") + markdownExtension
		}
		filename = filepath.Join(dir, cpath[1], basename)
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.WriteString(f, filePrepender(filename)); err != nil {
		return err
	}
	if err := GenMarkdownCustom(cmd, f, linkHandler); err != nil {
		return err
	}
	return nil
}
