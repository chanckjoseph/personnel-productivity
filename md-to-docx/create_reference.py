from docx import Document
from docx.shared import Pt, RGBColor
from docx.enum.style import WD_STYLE_TYPE

def create_styled_reference():
    doc = Document()

    # 1. Modify "Normal" style (Body Text)
    # Most body text uses 'Normal' or 'Body Text'
    style = doc.styles['Normal']
    font = style.font
    font.name = 'Arial'
    font.size = Pt(11)
    
    # 2. Modify "Heading 1"
    # Word usually has 'Heading 1' built-in, we just access it
    h1 = doc.styles['Heading 1']
    h1_font = h1.font
    h1_font.name = 'Arial'
    h1_font.size = Pt(24)
    h1_font.color.rgb = RGBColor(0, 51, 102) # Dark Blue
    h1_font.bold = True
    
    # 3. Modify "Heading 2"
    h2 = doc.styles['Heading 2']
    h2_font = h2.font
    h2_font.name = 'Arial'
    h2_font.size = Pt(18)
    h2_font.color.rgb = RGBColor(0, 85, 164) # Medium Blue
    h2_font.bold = True

    # 4. Modify "Heading 3"
    h3 = doc.styles['Heading 3']
    h3_font = h3.font
    h3_font.name = 'Arial'
    h3_font.size = Pt(14)
    h3_font.color.rgb = RGBColor(0, 51, 102) # Dark Blue
    h3_font.bold = True

    # 5. Modify "Code" or "Verbatim" style (for code blocks)
    # Pandoc often maps code blocks to a style called 'Source Code' or uses 'Plain Text' with a monospaced font
    # Let's try to create/modify a style for that if possible, but Pandoc's mapping is specific.
    # Usually it's better to rely on highlighting.
    
    doc.save('reference.docx')
    print("Created 'reference.docx' with custom Arial / Blue styles.")

if __name__ == "__main__":
    create_styled_reference()