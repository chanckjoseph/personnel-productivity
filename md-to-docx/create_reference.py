from docx import Document
from docx.shared import Pt, RGBColor
from docx.enum.style import WD_STYLE_TYPE
from pathlib import Path

from docx.oxml import OxmlElement
from docx.oxml.ns import qn

def set_table_style_borders(style):
    """
    Force borders on a table style using low-level XML manipulation.
    """
    if not style:
        return

    # Get or create tblPr
    # CT_Style has w:tblPr child
    tbl_pr = style._element.find(qn('w:tblPr'))
    if tbl_pr is None:
        tbl_pr = OxmlElement('w:tblPr')
        # We should insert it at the correct position, but appending might work for now
        # or we check if there are other known children.
        style._element.insert(0, tbl_pr) # Actually, order matters. styles usually have name, aliases, etc first.
                                         # But inserting at end is safer than 0 if we don't know. 
                                         # However, Schema says: name?, aliases?, ..., pPr?, rPr?, tblPr?, ...
                                         # Let's just append and hope python-docx handles it or Word ignores order.
        style._element.append(tbl_pr)

    # Get or create tblBorders
    tbl_borders = tbl_pr.find(qn('w:tblBorders'))
    if tbl_borders is None:
        tbl_borders = OxmlElement('w:tblBorders')
        tbl_pr.append(tbl_borders)

    # Define border types: top, left, bottom, right, insideH, insideV
    for border_name in ['top', 'left', 'bottom', 'right', 'insideH', 'insideV']:
        border = tbl_borders.find(qn(f'w:{border_name}'))
        if border is None:
            border = OxmlElement(f'w:{border_name}')
            tbl_borders.append(border)
        
        # Set attributes: single line, size 4 (1/2 pt), color auto, space 0
        border.set(qn('w:val'), 'single')
        border.set(qn('w:sz'), '4') 
        border.set(qn('w:space'), '0')
        border.set(qn('w:color'), 'auto')

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

    # 5. Modify Table Styles
    # We will FORCE the "Table Grid" style to be the default for all tables if possible,
    # or just ensure "Normal Table" (which is the default) has borders.
    
    # 5a. Configure "Normal Table" (The default style)
    try:
        normal_table = doc.styles['Normal Table']
        normal_table.font.name = 'Arial'
        normal_table.font.size = Pt(10)
        # Force borders on "Normal Table"
        set_table_style_borders(normal_table)
    except KeyError:
        print("'Normal Table' style not found.")

    # 5b. Configure "Table Grid" (standard bordered table)
    try:
        table_style = doc.styles['Table Grid']
        table_font = table_style.font
        table_font.name = 'Arial'
        table_font.size = Pt(10)
        set_table_style_borders(table_style)
    except KeyError:
        print("'Table Grid' style not found.")
        
    # Create a brand new style "MyTableStyle" based on "Table Grid"
    # This is hard with python-docx directly.
    # But we can try to use LatentStyles? No.
    
    # EASIER: Modify "Table Grid" but ensure we can reference it.
    # It seems "Table Grid" is built-in.
    
    # Try adding a stylealias?
    
    # OR: force "Normal Table" to work by ensuring the document defaults don't override it.
    
    # Let's try adding a style "MyCustomTable" if possible?
    try:
        # Styles.add_style(name, type, builtin=False)
        # Type for table is 3
        new_style = doc.styles.add_style('MyCustomTable', WD_STYLE_TYPE.TABLE)
        new_style.base_style = doc.styles['Table Grid']
        new_style.font.name = 'Arial'
        new_style.font.size = Pt(10)
        set_table_style_borders(new_style)
        print("Created 'MyCustomTable' style.")
    except Exception as e:
        print(f"Error creating custom style: {e}")

    output_path = Path("reference.docx")
    if not output_path.exists():
        # Try parent dir if running inside md-to-docx
        parent_path = Path("../reference.docx")
        output_path = parent_path
        
    doc.save(str(output_path))
    print(f"Created '{output_path}' with custom Arial / Blue styles.")

if __name__ == "__main__":
    create_styled_reference()
