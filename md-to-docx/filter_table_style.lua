function Table(el)
  -- Use "MyCustomTable" style
  -- Classes mapping to style names usually works best if style name has no spaces.
  
  -- Clear all classes and set MyCustomTable
  el.classes = pandoc.List({'MyCustomTable'})
  
  -- Also set custom-style attribute just in case
  el.attributes = {} 
  el.attributes['custom-style'] = 'MyCustomTable'
  
  -- Try setting the Attr object explicitely to avoid any existing attr issues
  -- el.attr = pandoc.Attr("", {"MyCustomTable"}, {["custom-style"] = "MyCustomTable"})
  
  return el
end
