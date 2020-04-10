package md

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/gomarkdown/markdown/ast"
)

// Renderer renders to markdown. Allows to convert to a canonnical
// form
type Renderer struct {
	orderedListCounter map[int]int
	// used to keep track of whether a given list item uses a paragraph
	// for large spacing.
	paragraph map[int]bool

	lastOutputLen  int
	listDepth      int
	lastNormalText string
}

// NewRenderer returns a Markdown renderer.
func NewRenderer() *Renderer {
	return &Renderer{
		orderedListCounter: map[int]int{},
		paragraph:          map[int]bool{},
	}
}

func (r *Renderer) out(w io.Writer, d []byte) {
	r.lastOutputLen = len(d)
	w.Write(d)
}

func (r *Renderer) outs(w io.Writer, s string) {
	r.lastOutputLen = len(s)
	io.WriteString(w, s)
}

func (r *Renderer) cr(w io.Writer) {
	if r.lastOutputLen > 0 {
		r.outs(w, "\n")
	}
}

func (r *Renderer) doubleSpace(w io.Writer) {
	// TODO: need to remember number of written bytes
	//if out.Len() > 0 {
	r.outs(w, "\n")
	//}
}

func (r *Renderer) list(w io.Writer, node *ast.List, entering bool) {
	if entering {
		r.listDepth++
		flags := node.ListFlags
		if flags&ast.ListTypeOrdered != 0 {
			r.orderedListCounter[r.listDepth] = 1
		}
	} else {
		r.listDepth--
	}

}

func (r *Renderer) heading(w io.Writer, node *ast.Heading, entering bool) {
	if entering {
		//if level >= 3 {
		r.outs(w, strings.Repeat("#", node.Level))
		r.outs(w, " ")
		//}
	} else {
		r.outs(w, "\n\n")
		/*
			l = 10
			switch level {
			case 1:
				r.outs(w, strings.Repeat("=", l))
				r.outs(w, "\n")
			case 2:
				r.outs(w, strings.Repeat("-", l))
				r.outs(w, "\n")
			}
		*/
	}
}

func (r *Renderer) listItem(w io.Writer, node *ast.ListItem, entering bool) {
	flags := node.ListFlags
	text := node.Literal

	if entering {
		if flags&ast.ListTypeOrdered != 0 {
			fmt.Fprintf(w, "%d. %s", r.orderedListCounter[r.listDepth], text)
			r.orderedListCounter[r.listDepth]++
		} else {
			//r.outs(w, "-")
			//indentwriter.New(w, 1).Write(text)
			fmt.Fprintf(w, "- %s", text)
		}
	} else {
		if r.paragraph[r.listDepth] {
			if flags&ast.ListItemEndOfList == 0 {
				// TODO why this prints each line as end of list?!
				//r.outs(w, "\n")
			}
			r.paragraph[r.listDepth] = false
		}
	}
}

func (r *Renderer) para(w io.Writer, node *ast.Paragraph, entering bool) {
	if entering {
		// marker := out.Len()
		//r.doubleSpace(w)

		r.paragraph[r.listDepth] = true
		r.outs(w, " ")
	} else {
		r.cr(w)
	}
}

// escape replaces instances of backslash with escaped backslash in text.
func escape(text []byte) []byte {
	return bytes.Replace(text, []byte(`\`), []byte(`\\`), -1)
}

func isNumber(data []byte) bool {
	for _, b := range data {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

func needsEscaping(text []byte, lastNormalText string) bool {
	switch string(text) {
	case `\`,
		"`",
		"*",
		"_",
		"{", "}",
		"[", "]",
		"(", ")",
		"#",
		"+",
		"-":
		return true
	case "!":
		return false
	case ".":
		// Return true if number, because a period after a number must be escaped to not get parsed as an ordered list.
		return isNumber([]byte(lastNormalText))
	case "<", ">":
		return true
	default:
		return false
	}
}

// cleanWithoutTrim is like clean, but doesn't trim blanks.
func cleanWithoutTrim(s string) string {
	var b []byte
	var p byte
	for i := 0; i < len(s); i++ {
		q := s[i]
		if q == '\n' || q == '\r' || q == '\t' {
			q = ' '
		}
		if q != ' ' || p != ' ' {
			b = append(b, q)
			p = q
		}
	}
	return string(b)
}

func (r *Renderer) skipSpaceIfNeededNormalText(w io.Writer, cleanString string) bool {
	if cleanString[0] != ' ' {
		return false
	}

	return false
	//  TODO: what did it mean to do?
	// we no longer use *bytes.Buffer for out, so whatever this tracked,
	// it has to be done in a different wy
	/*
		if _, ok := r.normalTextMarker[out]; !ok {
			r.normalTextMarker[out] = -1
		}
		return r.normalTextMarker[out] == out.Len()
	*/
}

func (r *Renderer) text(w io.Writer, text *ast.Text) {
	lit := text.Literal
	normalText := string(text.Literal)
	if needsEscaping(lit, r.lastNormalText) {
		lit = append([]byte("\\"), lit...)
	}
	r.lastNormalText = normalText
	if r.listDepth > 0 && string(lit) == "\n" {
		// TODO: See if this can be cleaned up... It's needed for lists.
		return
	}
	cleanString := cleanWithoutTrim(string(lit))
	if cleanString == "" {
		return
	}
	// Skip first space if last character is already a space (i.e., no need for a 2nd space in a row).
	if r.skipSpaceIfNeededNormalText(w, cleanString) {
		cleanString = cleanString[1:]
	}
	r.outs(w, cleanString)
	// If it ends with a space, make note of that.
	//if len(cleanString) >= 1 && cleanString[len(cleanString)-1] == ' ' {
	// TODO: write equivalent of this
	// r.normalTextMarker[out] = out.Len()
	//}
}

func (r *Renderer) del(w io.Writer, node *ast.Del) {
	r.outs(w, "~~")
	r.out(w, node.Literal)
	r.outs(w, "~~")
}

func (r *Renderer) strong(w io.Writer, node *ast.Strong) {
	text := node.Literal
	r.outs(w, "**")
	r.out(w, text)
	r.outs(w, "**")
}

func (r *Renderer) emphasis(w io.Writer, node *ast.Emph) {
	text := node.Literal
	if len(text) == 0 {
		return
	}
	r.outs(w, "*")
	r.out(w, text)
	r.outs(w, "*")
}

func (r *Renderer) htmlSpan(w io.Writer, node *ast.HTMLSpan) {
	r.out(w, node.Literal)
}

func (r *Renderer) htmlBlock(w io.Writer, node *ast.HTMLBlock) {
	r.doubleSpace(w)
	r.out(w, node.Literal)
	r.outs(w, "\n")
}

func (r *Renderer) codeBlock(w io.Writer, node *ast.CodeBlock) {
	r.doubleSpace(w)
	text := node.Literal
	lang := string(node.Info)
	// Parse out the language name.
	count := 0
	for _, elt := range strings.Fields(lang) {
		if elt[0] == '.' {
			elt = elt[1:]
		}
		if len(elt) == 0 {
			continue
		}
		r.outs(w, "```")
		r.outs(w, elt)
		count++
		break
	}

	if count == 0 {
		r.outs(w, "```")
	}
	r.outs(w, "\n")
	r.out(w, text)
	r.outs(w, "```\n")
}

func (r *Renderer) code(w io.Writer, node *ast.Code) {
	r.outs(w, "`")
	r.out(w, node.Literal)
	r.outs(w, "`")
}

func (r *Renderer) image(w io.Writer, node *ast.Image) {
	link := node.Destination
	title := node.Title
	// alt := node. ??
	var alt []byte
	r.outs(w, "![")
	r.out(w, alt)
	r.outs(w, "](")
	r.out(w, escape(link))
	if len(title) != 0 {
		r.outs(w, ` "`)
		r.out(w, title)
		r.outs(w, `"`)
	}
	r.outs(w, ")")
}

func (r *Renderer) link(w io.Writer, node *ast.Link) {
	link := string(escape(node.Destination))
	title := string(node.Title)
	content := string(node.Literal)
	r.outs(w, "[")
	r.outs(w, content)
	r.outs(w, "](")
	r.outs(w, link)
	if len(title) != 0 {
		r.outs(w, ` "`)
		r.outs(w, title)
		r.outs(w, `"`)
	}
	r.outs(w, ")")
}

// RenderNode renders markdown node
func (r *Renderer) RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	switch node := node.(type) {
	case *ast.Text:
		r.text(w, node)
	case *ast.Softbreak:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.Hardbreak:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.Emph:
		r.emphasis(w, node)
	case *ast.Strong:
		r.strong(w, node)
	case *ast.Del:
		r.del(w, node)
	case *ast.BlockQuote:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.Aside:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.Link:
		r.link(w, node)
	case *ast.CrossReference:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.Citation:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.Image:
		r.image(w, node)
	case *ast.Code:
		r.code(w, node)
	case *ast.CodeBlock:
		r.codeBlock(w, node)
	case *ast.Caption:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.CaptionFigure:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.Document:
		// do nothing
	case *ast.Paragraph:
		r.para(w, node, entering)
	case *ast.HTMLSpan:
		r.htmlSpan(w, node)
	case *ast.HTMLBlock:
		r.htmlBlock(w, node)
	case *ast.Heading:
		r.heading(w, node, entering)
	case *ast.HorizontalRule:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.List:
		r.list(w, node, entering)
	case *ast.ListItem:
		r.listItem(w, node, entering)
	case *ast.Table:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.TableCell:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.TableHeader:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.TableBody:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.TableRow:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.TableFooter:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.Math:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.MathBlock:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.DocumentMatter:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.Callout:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.Index:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.Subscript:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.Superscript:
		panic(fmt.Sprintf("node %T NYI", node))
	case *ast.Footnotes:
		// nothing by default; just output the list.
	default:
		panic(fmt.Sprintf("Unknown node %T", node))
	}
	return ast.GoToNext
}

// RenderHeader renders header
func (r *Renderer) RenderHeader(w io.Writer, ast ast.Node) {
	// do nothing
}

// RenderFooter renders footer
func (r *Renderer) RenderFooter(w io.Writer, ast ast.Node) {
	// do nothing
}
