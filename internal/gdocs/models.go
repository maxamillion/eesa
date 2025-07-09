package gdocs

import (
	"time"
)

// CreateDocumentRequest represents a request to create a new document
type CreateDocumentRequest struct {
	Title string `json:"title"`
}

// BatchUpdateDocumentRequest represents a batch update request
type BatchUpdateDocumentRequest struct {
	Requests       []Request `json:"requests"`
	WriteControl   *WriteControl `json:"writeControl,omitempty"`
}

// WriteControl represents write control for document updates
type WriteControl struct {
	RequiredRevisionID string `json:"requiredRevisionId,omitempty"`
}

// Request represents a single update request
type Request struct {
	InsertText               *InsertTextRequest               `json:"insertText,omitempty"`
	DeleteContentRange       *DeleteContentRangeRequest       `json:"deleteContentRange,omitempty"`
	UpdateTextStyle          *UpdateTextStyleRequest          `json:"updateTextStyle,omitempty"`
	UpdateParagraphStyle     *UpdateParagraphStyleRequest     `json:"updateParagraphStyle,omitempty"`
	CreateParagraphBullets   *CreateParagraphBulletsRequest   `json:"createParagraphBullets,omitempty"`
	DeleteParagraphBullets   *DeleteParagraphBulletsRequest   `json:"deleteParagraphBullets,omitempty"`
	InsertTable              *InsertTableRequest              `json:"insertTable,omitempty"`
	InsertTableRow           *InsertTableRowRequest           `json:"insertTableRow,omitempty"`
	InsertTableColumn        *InsertTableColumnRequest        `json:"insertTableColumn,omitempty"`
	DeleteTableRow           *DeleteTableRowRequest           `json:"deleteTableRow,omitempty"`
	DeleteTableColumn        *DeleteTableColumnRequest        `json:"deleteTableColumn,omitempty"`
	InsertInlineImage        *InsertInlineImageRequest        `json:"insertInlineImage,omitempty"`
	ReplaceAllText           *ReplaceAllTextRequest           `json:"replaceAllText,omitempty"`
	CreateNamedRange         *CreateNamedRangeRequest         `json:"createNamedRange,omitempty"`
	DeleteNamedRange         *DeleteNamedRangeRequest         `json:"deleteNamedRange,omitempty"`
	InsertPageBreak          *InsertPageBreakRequest          `json:"insertPageBreak,omitempty"`
	UpdateDocumentStyle      *UpdateDocumentStyleRequest      `json:"updateDocumentStyle,omitempty"`
}

// InsertTextRequest represents a request to insert text
type InsertTextRequest struct {
	Text     string    `json:"text"`
	Location *Location `json:"location"`
}

// DeleteContentRangeRequest represents a request to delete content
type DeleteContentRangeRequest struct {
	Range *Range `json:"range"`
}

// UpdateTextStyleRequest represents a request to update text style
type UpdateTextStyleRequest struct {
	Range     *Range     `json:"range"`
	TextStyle *TextStyle `json:"textStyle"`
	Fields    string     `json:"fields"`
}

// UpdateParagraphStyleRequest represents a request to update paragraph style
type UpdateParagraphStyleRequest struct {
	Range          *Range          `json:"range"`
	ParagraphStyle *ParagraphStyle `json:"paragraphStyle"`
	Fields         string          `json:"fields"`
}

// CreateParagraphBulletsRequest represents a request to create paragraph bullets
type CreateParagraphBulletsRequest struct {
	Range      *Range      `json:"range"`
	BulletPreset string    `json:"bulletPreset,omitempty"`
}

// DeleteParagraphBulletsRequest represents a request to delete paragraph bullets
type DeleteParagraphBulletsRequest struct {
	Range *Range `json:"range"`
}

// InsertTableRequest represents a request to insert a table
type InsertTableRequest struct {
	Rows     int32     `json:"rows"`
	Columns  int32     `json:"columns"`
	Location *Location `json:"location"`
}

// InsertTableRowRequest represents a request to insert a table row
type InsertTableRowRequest struct {
	TableCellLocation *TableCellLocation `json:"tableCellLocation"`
	InsertBelow       bool               `json:"insertBelow"`
}

// InsertTableColumnRequest represents a request to insert a table column
type InsertTableColumnRequest struct {
	TableCellLocation *TableCellLocation `json:"tableCellLocation"`
	InsertRight       bool               `json:"insertRight"`
}

// DeleteTableRowRequest represents a request to delete a table row
type DeleteTableRowRequest struct {
	TableCellLocation *TableCellLocation `json:"tableCellLocation"`
}

// DeleteTableColumnRequest represents a request to delete a table column
type DeleteTableColumnRequest struct {
	TableCellLocation *TableCellLocation `json:"tableCellLocation"`
}

// InsertInlineImageRequest represents a request to insert an inline image
type InsertInlineImageRequest struct {
	URI      string    `json:"uri"`
	Location *Location `json:"location"`
	ObjectSize *Size   `json:"objectSize,omitempty"`
}

// ReplaceAllTextRequest represents a request to replace all text
type ReplaceAllTextRequest struct {
	ContainsText    *SubstringMatchCriteria `json:"containsText"`
	ReplaceText     string                  `json:"replaceText"`
	TabsCriteria    *TabsCriteria           `json:"tabsCriteria,omitempty"`
}

// CreateNamedRangeRequest represents a request to create a named range
type CreateNamedRangeRequest struct {
	Name  string `json:"name"`
	Range *Range `json:"range"`
}

// DeleteNamedRangeRequest represents a request to delete a named range
type DeleteNamedRangeRequest struct {
	Name    string `json:"name,omitempty"`
	NamedRangeId string `json:"namedRangeId,omitempty"`
}

// InsertPageBreakRequest represents a request to insert a page break
type InsertPageBreakRequest struct {
	Location *Location `json:"location"`
}

// UpdateDocumentStyleRequest represents a request to update document style
type UpdateDocumentStyleRequest struct {
	DocumentStyle *DocumentStyle `json:"documentStyle"`
	Fields        string         `json:"fields"`
}

// Location represents a location in a document
type Location struct {
	Index    int32 `json:"index"`
	TabId    string `json:"tabId,omitempty"`
}

// Range represents a range in a document
type Range struct {
	StartIndex int32  `json:"startIndex"`
	EndIndex   int32  `json:"endIndex"`
	TabId      string `json:"tabId,omitempty"`
}

// TextStyle represents text formatting styles
type TextStyle struct {
	Bold                 *bool            `json:"bold,omitempty"`
	Italic               *bool            `json:"italic,omitempty"`
	Underline            *bool            `json:"underline,omitempty"`
	Strikethrough        *bool            `json:"strikethrough,omitempty"`
	SmallCaps            *bool            `json:"smallCaps,omitempty"`
	BackgroundColor      *OptionalColor   `json:"backgroundColor,omitempty"`
	ForegroundColor      *OptionalColor   `json:"foregroundColor,omitempty"`
	FontSize             *Dimension       `json:"fontSize,omitempty"`
	WeightedFontFamily   *WeightedFontFamily `json:"weightedFontFamily,omitempty"`
	BaselineOffset       string           `json:"baselineOffset,omitempty"`
	Link                 *Link            `json:"link,omitempty"`
}

// ParagraphStyle represents paragraph formatting styles
type ParagraphStyle struct {
	HeadingId              string              `json:"headingId,omitempty"`
	NamedStyleType         string              `json:"namedStyleType,omitempty"`
	Alignment              string              `json:"alignment,omitempty"`
	LineSpacing            *float64            `json:"lineSpacing,omitempty"`
	Direction              string              `json:"direction,omitempty"`
	SpacingMode            string              `json:"spacingMode,omitempty"`
	SpaceAbove             *Dimension          `json:"spaceAbove,omitempty"`
	SpaceBelow             *Dimension          `json:"spaceBelow,omitempty"`
	BorderBetween          *ParagraphBorder    `json:"borderBetween,omitempty"`
	BorderTop              *ParagraphBorder    `json:"borderTop,omitempty"`
	BorderBottom           *ParagraphBorder    `json:"borderBottom,omitempty"`
	BorderLeft             *ParagraphBorder    `json:"borderLeft,omitempty"`
	BorderRight            *ParagraphBorder    `json:"borderRight,omitempty"`
	IndentFirstLine        *Dimension          `json:"indentFirstLine,omitempty"`
	IndentStart            *Dimension          `json:"indentStart,omitempty"`
	IndentEnd              *Dimension          `json:"indentEnd,omitempty"`
	TabStops               []TabStop           `json:"tabStops,omitempty"`
	KeepLinesTogether      *bool               `json:"keepLinesTogether,omitempty"`
	KeepWithNext           *bool               `json:"keepWithNext,omitempty"`
	AvoidWidowAndOrphan    *bool               `json:"avoidWidowAndOrphan,omitempty"`
	Shading                *Shading            `json:"shading,omitempty"`
	PageBreakBefore        *bool               `json:"pageBreakBefore,omitempty"`
}

// DocumentStyle represents document-level styles
type DocumentStyle struct {
	Background              *Background    `json:"background,omitempty"`
	PageNumberStart         *int32         `json:"pageNumberStart,omitempty"`
	MarginTop               *Dimension     `json:"marginTop,omitempty"`
	MarginBottom            *Dimension     `json:"marginBottom,omitempty"`
	MarginRight             *Dimension     `json:"marginRight,omitempty"`
	MarginLeft              *Dimension     `json:"marginLeft,omitempty"`
	PageSize                *Size          `json:"pageSize,omitempty"`
	MarginHeader            *Dimension     `json:"marginHeader,omitempty"`
	MarginFooter            *Dimension     `json:"marginFooter,omitempty"`
	UseCustomHeaderFooterMargins *bool     `json:"useCustomHeaderFooterMargins,omitempty"`
	EvenPageHeaderId        string         `json:"evenPageHeaderId,omitempty"`
	EvenPageFooterId        string         `json:"evenPageFooterId,omitempty"`
	FirstPageHeaderId       string         `json:"firstPageHeaderId,omitempty"`
	FirstPageFooterId       string         `json:"firstPageFooterId,omitempty"`
	DefaultHeaderId         string         `json:"defaultHeaderId,omitempty"`
	DefaultFooterId         string         `json:"defaultFooterId,omitempty"`
	UseEvenPageHeaderFooter *bool          `json:"useEvenPageHeaderFooter,omitempty"`
	UseFirstPageHeaderFooter *bool         `json:"useFirstPageHeaderFooter,omitempty"`
	FlipPageOrientation     *bool          `json:"flipPageOrientation,omitempty"`
}

// TableCellLocation represents a location in a table cell
type TableCellLocation struct {
	TableStartLocation *Location `json:"tableStartLocation"`
	RowIndex           int32     `json:"rowIndex"`
	ColumnIndex        int32     `json:"columnIndex"`
}

// SubstringMatchCriteria represents criteria for substring matching
type SubstringMatchCriteria struct {
	Text         string `json:"text"`
	MatchCase    *bool  `json:"matchCase,omitempty"`
}

// TabsCriteria represents criteria for tabs
type TabsCriteria struct {
	TabIds []string `json:"tabIds"`
}

// OptionalColor represents an optional color
type OptionalColor struct {
	Color *Color `json:"color,omitempty"`
}

// Color represents a color
type Color struct {
	RgbColor *RgbColor `json:"rgbColor,omitempty"`
}

// RgbColor represents an RGB color
type RgbColor struct {
	Red   float64 `json:"red"`
	Green float64 `json:"green"`
	Blue  float64 `json:"blue"`
}

// Dimension represents a dimension with magnitude and unit
type Dimension struct {
	Magnitude float64 `json:"magnitude"`
	Unit      string  `json:"unit"`
}

// Size represents a size with width and height
type Size struct {
	Width  *Dimension `json:"width,omitempty"`
	Height *Dimension `json:"height,omitempty"`
}

// WeightedFontFamily represents a font family with weight
type WeightedFontFamily struct {
	FontFamily string `json:"fontFamily"`
	Weight     int32  `json:"weight,omitempty"`
}

// Link represents a hyperlink
type Link struct {
	URL        string `json:"url,omitempty"`
	BookmarkId string `json:"bookmarkId,omitempty"`
	HeadingId  string `json:"headingId,omitempty"`
	TabId      string `json:"tabId,omitempty"`
}

// ParagraphBorder represents a paragraph border
type ParagraphBorder struct {
	Color       *OptionalColor `json:"color,omitempty"`
	Width       *Dimension     `json:"width,omitempty"`
	Padding     *Dimension     `json:"padding,omitempty"`
	DashStyle   string         `json:"dashStyle,omitempty"`
}

// TabStop represents a tab stop
type TabStop struct {
	Offset    *Dimension `json:"offset"`
	Alignment string     `json:"alignment"`
}

// Shading represents paragraph shading
type Shading struct {
	BackgroundColor *OptionalColor `json:"backgroundColor,omitempty"`
}

// Background represents document background
type Background struct {
	Color *OptionalColor `json:"color,omitempty"`
}

// DocumentResponse represents a Google Docs document
type DocumentResponse struct {
	DocumentID     string               `json:"documentId"`
	Title          string               `json:"title"`
	Body           *Body                `json:"body,omitempty"`
	Headers        map[string]*Header   `json:"headers,omitempty"`
	Footers        map[string]*Footer   `json:"footers,omitempty"`
	DocumentStyle  *DocumentStyle       `json:"documentStyle,omitempty"`
	NamedStyles    *NamedStyles         `json:"namedStyles,omitempty"`
	RevisionID     string               `json:"revisionId,omitempty"`
	SuggestionsViewMode string          `json:"suggestionsViewMode,omitempty"`
	NamedRanges    map[string]*NamedRanges `json:"namedRanges,omitempty"`
	Lists          map[string]*List     `json:"lists,omitempty"`
	InlineObjects  map[string]*InlineObject `json:"inlineObjects,omitempty"`
	Tabs           []Tab                `json:"tabs,omitempty"`
}

// Body represents the document body
type Body struct {
	Content []StructuralElement `json:"content"`
}

// Header represents a document header
type Header struct {
	HeaderID string                `json:"headerId"`
	Content  []StructuralElement   `json:"content"`
}

// Footer represents a document footer
type Footer struct {
	FooterID string                `json:"footerId"`
	Content  []StructuralElement   `json:"content"`
}

// NamedStyles represents named styles in the document
type NamedStyles struct {
	Styles []NamedStyle `json:"styles"`
}

// NamedStyle represents a named style
type NamedStyle struct {
	NamedStyleType string          `json:"namedStyleType"`
	TextStyle      *TextStyle      `json:"textStyle,omitempty"`
	ParagraphStyle *ParagraphStyle `json:"paragraphStyle,omitempty"`
}

// NamedRanges represents named ranges
type NamedRanges struct {
	Name       string       `json:"name"`
	NamedRanges []NamedRange `json:"namedRanges"`
}

// NamedRange represents a named range
type NamedRange struct {
	NamedRangeID string `json:"namedRangeId"`
	Name         string `json:"name"`
	Ranges       []Range `json:"ranges"`
}

// List represents a list in the document
type List struct {
	ListID               string                          `json:"listId"`
	ListProperties       *ListProperties                 `json:"listProperties,omitempty"`
	SuggestedInsertionIds []string                       `json:"suggestedInsertionIds,omitempty"`
	SuggestedDeletionIds []string                       `json:"suggestedDeletionIds,omitempty"`
}

// ListProperties represents list properties
type ListProperties struct {
	NestingLevels []NestingLevel `json:"nestingLevels"`
}

// NestingLevel represents a nesting level in a list
type NestingLevel struct {
	BulletAlignment string       `json:"bulletAlignment,omitempty"`
	GlyphFormat     string       `json:"glyphFormat,omitempty"`
	GlyphSymbol     string       `json:"glyphSymbol,omitempty"`
	GlyphType       string       `json:"glyphType,omitempty"`
	IndentFirstLine *Dimension   `json:"indentFirstLine,omitempty"`
	IndentStart     *Dimension   `json:"indentStart,omitempty"`
	StartNumber     int32        `json:"startNumber,omitempty"`
	TextStyle       *TextStyle   `json:"textStyle,omitempty"`
}

// InlineObject represents an inline object
type InlineObject struct {
	ObjectID               string                        `json:"objectId"`
	InlineObjectProperties *InlineObjectProperties       `json:"inlineObjectProperties,omitempty"`
	SuggestedInsertionIds  []string                     `json:"suggestedInsertionIds,omitempty"`
	SuggestedDeletionIds   []string                     `json:"suggestedDeletionIds,omitempty"`
}

// InlineObjectProperties represents inline object properties
type InlineObjectProperties struct {
	EmbeddedObject *EmbeddedObject `json:"embeddedObject,omitempty"`
}

// EmbeddedObject represents an embedded object
type EmbeddedObject struct {
	Title                string            `json:"title,omitempty"`
	Description          string            `json:"description,omitempty"`
	EmbeddedObjectBorder *EmbeddedObjectBorder `json:"embeddedObjectBorder,omitempty"`
	Size                 *Size             `json:"size,omitempty"`
	MarginTop            *Dimension        `json:"marginTop,omitempty"`
	MarginBottom         *Dimension        `json:"marginBottom,omitempty"`
	MarginRight          *Dimension        `json:"marginRight,omitempty"`
	MarginLeft           *Dimension        `json:"marginLeft,omitempty"`
	LinkedContentReference *LinkedContentReference `json:"linkedContentReference,omitempty"`
	ImageProperties      *ImageProperties  `json:"imageProperties,omitempty"`
}

// EmbeddedObjectBorder represents an embedded object border
type EmbeddedObjectBorder struct {
	Color     *OptionalColor `json:"color,omitempty"`
	Width     *Dimension     `json:"width,omitempty"`
	DashStyle string         `json:"dashStyle,omitempty"`
}

// LinkedContentReference represents a linked content reference
type LinkedContentReference struct {
	SheetsChartReference *SheetsChartReference `json:"sheetsChartReference,omitempty"`
}

// SheetsChartReference represents a reference to a Sheets chart
type SheetsChartReference struct {
	SpreadsheetID string `json:"spreadsheetId"`
	ChartID       int32  `json:"chartId"`
}

// ImageProperties represents image properties
type ImageProperties struct {
	ContentURI   string     `json:"contentUri,omitempty"`
	SourceURI    string     `json:"sourceUri,omitempty"`
	Brightness   *float64   `json:"brightness,omitempty"`
	Contrast     *float64   `json:"contrast,omitempty"`
	Transparency *float64   `json:"transparency,omitempty"`
	CropProperties *CropProperties `json:"cropProperties,omitempty"`
}

// CropProperties represents image crop properties
type CropProperties struct {
	OffsetLeft   *float64 `json:"offsetLeft,omitempty"`
	OffsetRight  *float64 `json:"offsetRight,omitempty"`
	OffsetTop    *float64 `json:"offsetTop,omitempty"`
	OffsetBottom *float64 `json:"offsetBottom,omitempty"`
	Angle        *float64 `json:"angle,omitempty"`
}

// Tab represents a document tab
type Tab struct {
	TabID          string                `json:"tabId"`
	Index          int32                 `json:"index"`
	ChildTabs      []Tab                 `json:"childTabs,omitempty"`
	TabProperties  *TabProperties        `json:"tabProperties,omitempty"`
	DocumentTab    *DocumentTab          `json:"documentTab,omitempty"`
}

// TabProperties represents tab properties
type TabProperties struct {
	Title      string `json:"title,omitempty"`
	Index      int32  `json:"index,omitempty"`
	NestingLevel int32 `json:"nestingLevel,omitempty"`
}

// DocumentTab represents a document tab
type DocumentTab struct {
	Body          *Body          `json:"body,omitempty"`
	Headers       map[string]*Header `json:"headers,omitempty"`
	Footers       map[string]*Footer `json:"footers,omitempty"`
	DocumentStyle *DocumentStyle `json:"documentStyle,omitempty"`
}

// StructuralElement represents a structural element in the document
type StructuralElement struct {
	StartIndex        int32              `json:"startIndex"`
	EndIndex          int32              `json:"endIndex"`
	Paragraph         *Paragraph         `json:"paragraph,omitempty"`
	SectionBreak      *SectionBreak      `json:"sectionBreak,omitempty"`
	Table             *Table             `json:"table,omitempty"`
	TableOfContents   *TableOfContents   `json:"tableOfContents,omitempty"`
}

// Paragraph represents a paragraph
type Paragraph struct {
	Elements          []ParagraphElement `json:"elements"`
	ParagraphStyle    *ParagraphStyle    `json:"paragraphStyle,omitempty"`
	PositionedObjectIds []string         `json:"positionedObjectIds,omitempty"`
	SuggestedParagraphStyleChanges map[string]*SuggestedParagraphStyle `json:"suggestedParagraphStyleChanges,omitempty"`
	SuggestedPositionedObjectIds   map[string]*ObjectReferences        `json:"suggestedPositionedObjectIds,omitempty"`
	Bullet            *Bullet            `json:"bullet,omitempty"`
}

// ParagraphElement represents an element within a paragraph
type ParagraphElement struct {
	StartIndex               int32                     `json:"startIndex"`
	EndIndex                 int32                     `json:"endIndex"`
	TextRun                  *TextRun                  `json:"textRun,omitempty"`
	AutoText                 *AutoText                 `json:"autoText,omitempty"`
	PageBreak                *PageBreak                `json:"pageBreak,omitempty"`
	ColumnBreak              *ColumnBreak              `json:"columnBreak,omitempty"`
	FootnoteReference        *FootnoteReference        `json:"footnoteReference,omitempty"`
	HorizontalRule           *HorizontalRule           `json:"horizontalRule,omitempty"`
	Equation                 *Equation                 `json:"equation,omitempty"`
	InlineObjectElement      *InlineObjectElement      `json:"inlineObjectElement,omitempty"`
	Person                   *Person                   `json:"person,omitempty"`
	RichLink                 *RichLink                 `json:"richLink,omitempty"`
}

// TextRun represents a run of text
type TextRun struct {
	Content                string                          `json:"content"`
	TextStyle              *TextStyle                      `json:"textStyle,omitempty"`
	SuggestedInsertionIds  []string                        `json:"suggestedInsertionIds,omitempty"`
	SuggestedDeletionIds   []string                        `json:"suggestedDeletionIds,omitempty"`
	SuggestedTextStyleChanges map[string]*SuggestedTextStyle `json:"suggestedTextStyleChanges,omitempty"`
}

// AutoText represents auto text
type AutoText struct {
	Type     string     `json:"type"`
	TextStyle *TextStyle `json:"textStyle,omitempty"`
	SuggestedInsertionIds []string `json:"suggestedInsertionIds,omitempty"`
	SuggestedDeletionIds  []string `json:"suggestedDeletionIds,omitempty"`
	SuggestedTextStyleChanges map[string]*SuggestedTextStyle `json:"suggestedTextStyleChanges,omitempty"`
}

// PageBreak represents a page break
type PageBreak struct {
	TextStyle             *TextStyle                      `json:"textStyle,omitempty"`
	SuggestedInsertionIds []string                        `json:"suggestedInsertionIds,omitempty"`
	SuggestedDeletionIds  []string                        `json:"suggestedDeletionIds,omitempty"`
	SuggestedTextStyleChanges map[string]*SuggestedTextStyle `json:"suggestedTextStyleChanges,omitempty"`
}

// ColumnBreak represents a column break
type ColumnBreak struct {
	TextStyle             *TextStyle                      `json:"textStyle,omitempty"`
	SuggestedInsertionIds []string                        `json:"suggestedInsertionIds,omitempty"`
	SuggestedDeletionIds  []string                        `json:"suggestedDeletionIds,omitempty"`
	SuggestedTextStyleChanges map[string]*SuggestedTextStyle `json:"suggestedTextStyleChanges,omitempty"`
}

// FootnoteReference represents a footnote reference
type FootnoteReference struct {
	FootnoteID            string                          `json:"footnoteId"`
	FootnoteNumber        string                          `json:"footnoteNumber"`
	TextStyle             *TextStyle                      `json:"textStyle,omitempty"`
	SuggestedInsertionIds []string                        `json:"suggestedInsertionIds,omitempty"`
	SuggestedDeletionIds  []string                        `json:"suggestedDeletionIds,omitempty"`
	SuggestedTextStyleChanges map[string]*SuggestedTextStyle `json:"suggestedTextStyleChanges,omitempty"`
}

// HorizontalRule represents a horizontal rule
type HorizontalRule struct {
	TextStyle             *TextStyle                      `json:"textStyle,omitempty"`
	SuggestedInsertionIds []string                        `json:"suggestedInsertionIds,omitempty"`
	SuggestedDeletionIds  []string                        `json:"suggestedDeletionIds,omitempty"`
	SuggestedTextStyleChanges map[string]*SuggestedTextStyle `json:"suggestedTextStyleChanges,omitempty"`
}

// Equation represents an equation
type Equation struct {
	SuggestedInsertionIds []string `json:"suggestedInsertionIds,omitempty"`
	SuggestedDeletionIds  []string `json:"suggestedDeletionIds,omitempty"`
}

// InlineObjectElement represents an inline object element
type InlineObjectElement struct {
	InlineObjectID        string                          `json:"inlineObjectId"`
	TextStyle             *TextStyle                      `json:"textStyle,omitempty"`
	SuggestedInsertionIds []string                        `json:"suggestedInsertionIds,omitempty"`
	SuggestedDeletionIds  []string                        `json:"suggestedDeletionIds,omitempty"`
	SuggestedTextStyleChanges map[string]*SuggestedTextStyle `json:"suggestedTextStyleChanges,omitempty"`
}

// Person represents a person
type Person struct {
	PersonID              string                          `json:"personId"`
	PersonProperties      *PersonProperties               `json:"personProperties,omitempty"`
	TextStyle             *TextStyle                      `json:"textStyle,omitempty"`
	SuggestedInsertionIds []string                        `json:"suggestedInsertionIds,omitempty"`
	SuggestedDeletionIds  []string                        `json:"suggestedDeletionIds,omitempty"`
	SuggestedTextStyleChanges map[string]*SuggestedTextStyle `json:"suggestedTextStyleChanges,omitempty"`
}

// PersonProperties represents person properties
type PersonProperties struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// RichLink represents a rich link
type RichLink struct {
	RichLinkID            string                          `json:"richLinkId"`
	RichLinkProperties    *RichLinkProperties             `json:"richLinkProperties,omitempty"`
	TextStyle             *TextStyle                      `json:"textStyle,omitempty"`
	SuggestedInsertionIds []string                        `json:"suggestedInsertionIds,omitempty"`
	SuggestedDeletionIds  []string                        `json:"suggestedDeletionIds,omitempty"`
	SuggestedTextStyleChanges map[string]*SuggestedTextStyle `json:"suggestedTextStyleChanges,omitempty"`
}

// RichLinkProperties represents rich link properties
type RichLinkProperties struct {
	Title       string `json:"title"`
	URI         string `json:"uri"`
	MimeType    string `json:"mimeType"`
}

// SuggestedTextStyle represents suggested text style changes
type SuggestedTextStyle struct {
	TextStyle            *TextStyle `json:"textStyle,omitempty"`
	TextStyleSuggestionState *TextStyleSuggestionState `json:"textStyleSuggestionState,omitempty"`
}

// TextStyleSuggestionState represents text style suggestion state
type TextStyleSuggestionState struct {
	BaseTextStyle      *TextStyle `json:"baseTextStyle,omitempty"`
	SuggestedTextStyle *TextStyle `json:"suggestedTextStyle,omitempty"`
}

// SuggestedParagraphStyle represents suggested paragraph style changes
type SuggestedParagraphStyle struct {
	ParagraphStyle            *ParagraphStyle `json:"paragraphStyle,omitempty"`
	ParagraphStyleSuggestionState *ParagraphStyleSuggestionState `json:"paragraphStyleSuggestionState,omitempty"`
}

// ParagraphStyleSuggestionState represents paragraph style suggestion state
type ParagraphStyleSuggestionState struct {
	BaseParagraphStyle      *ParagraphStyle `json:"baseParagraphStyle,omitempty"`
	SuggestedParagraphStyle *ParagraphStyle `json:"suggestedParagraphStyle,omitempty"`
}

// ObjectReferences represents object references
type ObjectReferences struct {
	ObjectIds []string `json:"objectIds"`
}

// Bullet represents a bullet
type Bullet struct {
	ListID       string     `json:"listId"`
	NestingLevel int32      `json:"nestingLevel"`
	TextStyle    *TextStyle `json:"textStyle,omitempty"`
}

// SectionBreak represents a section break
type SectionBreak struct {
	SectionStyle          *SectionStyle `json:"sectionStyle,omitempty"`
	SuggestedInsertionIds []string      `json:"suggestedInsertionIds,omitempty"`
	SuggestedDeletionIds  []string      `json:"suggestedDeletionIds,omitempty"`
}

// SectionStyle represents section style
type SectionStyle struct {
	ColumnSeparatorStyle    string           `json:"columnSeparatorStyle,omitempty"`
	ContentDirection        string           `json:"contentDirection,omitempty"`
	SectionType             string           `json:"sectionType,omitempty"`
	EvenPageHeaderId        string           `json:"evenPageHeaderId,omitempty"`
	EvenPageFooterId        string           `json:"evenPageFooterId,omitempty"`
	FirstPageHeaderId       string           `json:"firstPageHeaderId,omitempty"`
	FirstPageFooterId       string           `json:"firstPageFooterId,omitempty"`
	DefaultHeaderId         string           `json:"defaultHeaderId,omitempty"`
	DefaultFooterId         string           `json:"defaultFooterId,omitempty"`
	UseFirstPageHeaderFooter *bool           `json:"useFirstPageHeaderFooter,omitempty"`
	PageNumberStart         *int32          `json:"pageNumberStart,omitempty"`
	MarginTop               *Dimension      `json:"marginTop,omitempty"`
	MarginBottom            *Dimension      `json:"marginBottom,omitempty"`
	MarginRight             *Dimension      `json:"marginRight,omitempty"`
	MarginLeft              *Dimension      `json:"marginLeft,omitempty"`
	PageSize                *Size           `json:"pageSize,omitempty"`
	MarginHeader            *Dimension      `json:"marginHeader,omitempty"`
	MarginFooter            *Dimension      `json:"marginFooter,omitempty"`
	FlipPageOrientation     *bool           `json:"flipPageOrientation,omitempty"`
	ColumnProperties        []ColumnProperties `json:"columnProperties,omitempty"`
}

// ColumnProperties represents column properties
type ColumnProperties struct {
	Width        *Dimension `json:"width,omitempty"`
	PaddingEnd   *Dimension `json:"paddingEnd,omitempty"`
}

// Table represents a table
type Table struct {
	Columns               int32            `json:"columns"`
	Rows                  int32            `json:"rows"`
	TableRows             []TableRow       `json:"tableRows"`
	TableStyle            *TableStyle      `json:"tableStyle,omitempty"`
	SuggestedInsertionIds []string         `json:"suggestedInsertionIds,omitempty"`
	SuggestedDeletionIds  []string         `json:"suggestedDeletionIds,omitempty"`
}

// TableRow represents a table row
type TableRow struct {
	StartIndex               int32             `json:"startIndex"`
	EndIndex                 int32             `json:"endIndex"`
	TableCells               []TableCell       `json:"tableCells"`
	TableRowStyle            *TableRowStyle    `json:"tableRowStyle,omitempty"`
	SuggestedInsertionIds    []string          `json:"suggestedInsertionIds,omitempty"`
	SuggestedDeletionIds     []string          `json:"suggestedDeletionIds,omitempty"`
	SuggestedTableRowStyleChanges map[string]*SuggestedTableRowStyle `json:"suggestedTableRowStyleChanges,omitempty"`
}

// TableCell represents a table cell
type TableCell struct {
	StartIndex               int32             `json:"startIndex"`
	EndIndex                 int32             `json:"endIndex"`
	Content                  []StructuralElement `json:"content"`
	TableCellStyle           *TableCellStyle   `json:"tableCellStyle,omitempty"`
	SuggestedInsertionIds    []string          `json:"suggestedInsertionIds,omitempty"`
	SuggestedDeletionIds     []string          `json:"suggestedDeletionIds,omitempty"`
	SuggestedTableCellStyleChanges map[string]*SuggestedTableCellStyle `json:"suggestedTableCellStyleChanges,omitempty"`
}

// TableStyle represents table style
type TableStyle struct {
	TableColumnProperties []TableColumnProperties `json:"tableColumnProperties,omitempty"`
}

// TableColumnProperties represents table column properties
type TableColumnProperties struct {
	Width        *Dimension `json:"width,omitempty"`
	WidthType    string     `json:"widthType,omitempty"`
}

// TableRowStyle represents table row style
type TableRowStyle struct {
	MinRowHeight              *Dimension `json:"minRowHeight,omitempty"`
	PreventOverflow           *bool      `json:"preventOverflow,omitempty"`
	TableHeader               *bool      `json:"tableHeader,omitempty"`
}

// TableCellStyle represents table cell style
type TableCellStyle struct {
	RowSpan                   int32              `json:"rowSpan,omitempty"`
	ColumnSpan                int32              `json:"columnSpan,omitempty"`
	BackgroundColor           *OptionalColor     `json:"backgroundColor,omitempty"`
	BorderTop                 *TableCellBorder   `json:"borderTop,omitempty"`
	BorderBottom              *TableCellBorder   `json:"borderBottom,omitempty"`
	BorderLeft                *TableCellBorder   `json:"borderLeft,omitempty"`
	BorderRight               *TableCellBorder   `json:"borderRight,omitempty"`
	PaddingTop                *Dimension         `json:"paddingTop,omitempty"`
	PaddingBottom             *Dimension         `json:"paddingBottom,omitempty"`
	PaddingLeft               *Dimension         `json:"paddingLeft,omitempty"`
	PaddingRight              *Dimension         `json:"paddingRight,omitempty"`
	ContentAlignment          string             `json:"contentAlignment,omitempty"`
}

// TableCellBorder represents a table cell border
type TableCellBorder struct {
	Color     *OptionalColor `json:"color,omitempty"`
	Width     *Dimension     `json:"width,omitempty"`
	DashStyle string         `json:"dashStyle,omitempty"`
}

// SuggestedTableRowStyle represents suggested table row style changes
type SuggestedTableRowStyle struct {
	TableRowStyle            *TableRowStyle `json:"tableRowStyle,omitempty"`
	TableRowStyleSuggestionState *TableRowStyleSuggestionState `json:"tableRowStyleSuggestionState,omitempty"`
}

// TableRowStyleSuggestionState represents table row style suggestion state
type TableRowStyleSuggestionState struct {
	BaseTableRowStyle      *TableRowStyle `json:"baseTableRowStyle,omitempty"`
	SuggestedTableRowStyle *TableRowStyle `json:"suggestedTableRowStyle,omitempty"`
}

// SuggestedTableCellStyle represents suggested table cell style changes
type SuggestedTableCellStyle struct {
	TableCellStyle            *TableCellStyle `json:"tableCellStyle,omitempty"`
	TableCellStyleSuggestionState *TableCellStyleSuggestionState `json:"tableCellStyleSuggestionState,omitempty"`
}

// TableCellStyleSuggestionState represents table cell style suggestion state
type TableCellStyleSuggestionState struct {
	BaseTableCellStyle      *TableCellStyle `json:"baseTableCellStyle,omitempty"`
	SuggestedTableCellStyle *TableCellStyle `json:"suggestedTableCellStyle,omitempty"`
}

// TableOfContents represents a table of contents
type TableOfContents struct {
	Content               []StructuralElement `json:"content"`
	SuggestedInsertionIds []string            `json:"suggestedInsertionIds,omitempty"`
	SuggestedDeletionIds  []string            `json:"suggestedDeletionIds,omitempty"`
}

// BatchUpdateResponse represents a batch update response
type BatchUpdateResponse struct {
	DocumentID string  `json:"documentId"`
	Replies    []Reply `json:"replies"`
}

// Reply represents a reply to a request
type Reply struct {
	CreateNamedRange   *CreateNamedRangeReply   `json:"createNamedRange,omitempty"`
	InsertText         *InsertTextReply         `json:"insertText,omitempty"`
	InsertInlineImage  *InsertInlineImageReply  `json:"insertInlineImage,omitempty"`
	InsertTable        *InsertTableReply        `json:"insertTable,omitempty"`
	ReplaceAllText     *ReplaceAllTextReply     `json:"replaceAllText,omitempty"`
}

// CreateNamedRangeReply represents a create named range reply
type CreateNamedRangeReply struct {
	NamedRangeID string `json:"namedRangeId"`
}

// InsertTextReply represents an insert text reply
type InsertTextReply struct {
	// This request type doesn't return any specific data
}

// InsertInlineImageReply represents an insert inline image reply
type InsertInlineImageReply struct {
	ObjectID string `json:"objectId"`
}

// InsertTableReply represents an insert table reply
type InsertTableReply struct {
	ObjectID string `json:"objectId"`
}

// ReplaceAllTextReply represents a replace all text reply
type ReplaceAllTextReply struct {
	OccurrencesChanged int32 `json:"occurrencesChanged"`
}

// Permission represents a permission for sharing
type Permission struct {
	Type         string `json:"type"`
	Role         string `json:"role"`
	EmailAddress string `json:"emailAddress,omitempty"`
	Domain       string `json:"domain,omitempty"`
	AllowFileDiscovery *bool `json:"allowFileDiscovery,omitempty"`
	ExpirationTime     string `json:"expirationTime,omitempty"`
}

// GoogleErrorResponse represents a Google API error response
type GoogleErrorResponse struct {
	ErrorInfo GoogleAPIError `json:"error"`
}

// GoogleAPIError represents a Google API error
type GoogleAPIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
	Details []GoogleErrorDetail `json:"details,omitempty"`
}

// GoogleErrorDetail represents Google API error details
type GoogleErrorDetail struct {
	Type     string `json:"@type"`
	Reason   string `json:"reason,omitempty"`
	Domain   string `json:"domain,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Error returns the error message from GoogleAPIError
func (e *GoogleAPIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "Unknown Google API error"
}

// Error returns the error message from GoogleErrorResponse
func (e *GoogleErrorResponse) Error() string {
	return e.ErrorInfo.Error()
}

// DocumentSummary represents a summary of document creation
type DocumentSummary struct {
	DocumentID     string                 `json:"documentId"`
	Title          string                 `json:"title"`
	URL            string                 `json:"url"`
	CreatedAt      time.Time              `json:"createdAt"`
	ContentLength  int                    `json:"contentLength"`
	SharedWith     []string               `json:"sharedWith"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// Constants for named style types
const (
	NamedStyleTypeNormalText    = "NORMAL_TEXT"
	NamedStyleTypeTitle         = "TITLE"
	NamedStyleTypeSubtitle      = "SUBTITLE"
	NamedStyleTypeHeading1      = "HEADING_1"
	NamedStyleTypeHeading2      = "HEADING_2"
	NamedStyleTypeHeading3      = "HEADING_3"
	NamedStyleTypeHeading4      = "HEADING_4"
	NamedStyleTypeHeading5      = "HEADING_5"
	NamedStyleTypeHeading6      = "HEADING_6"
)

// Constants for text alignment
const (
	AlignmentStart         = "START"
	AlignmentCenter        = "CENTER"
	AlignmentEnd           = "END"
	AlignmentJustify       = "JUSTIFY"
	AlignmentUnspecified   = "ALIGNMENT_UNSPECIFIED"
)

// Constants for bullet presets
const (
	BulletDiscCircleSquare       = "BULLET_DISC_CIRCLE_SQUARE"
	BulletDiamondCircleSquare    = "BULLET_DIAMONDX_CIRCLE_SQUARE"
	BulletCheckboxArrowDiamond   = "BULLET_CHECKBOX_ARROW_DIAMOND"
	BulletArrowDiamondDisc       = "BULLET_ARROW_DIAMOND_DISC"
	BulletStarCircleSquare       = "BULLET_STAR_CIRCLE_SQUARE"
	BulletArrow3DCircleSquare    = "BULLET_ARROW3D_CIRCLE_SQUARE"
	BulletLeftTriangleCircleSquare = "BULLET_LEFTTRIANGLE_CIRCLE_SQUARE"
	BulletDiamondArrowCircle     = "BULLET_DIAMOND_ARROW_CIRCLE"
	BulletCheckboxArrow3DCircle  = "BULLET_CHECKBOX_ARROW3D_CIRCLE"
	BulletCheckboxCircleSquare   = "BULLET_CHECKBOX_CIRCLE_SQUARE"
	BulletCheckboxSquareCircle   = "BULLET_CHECKBOX_SQUARE_CIRCLE"
)

// Constants for sharing roles
const (
	RoleOwner       = "owner"
	RoleOrganizer   = "organizer"
	RoleFileOrganizer = "fileOrganizer"
	RoleWriter      = "writer"
	RoleCommenter   = "commenter"
	RoleReader      = "reader"
)

// Constants for sharing types
const (
	TypeUser   = "user"
	TypeGroup  = "group"
	TypeDomain = "domain"
	TypeAnyone = "anyone"
)