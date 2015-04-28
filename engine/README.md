#Introduction
**engine** accepts a LaTex-like document and processes it with a set of functions \{f1,f2,f3...\}. Then it stores origianl text along with processesed test into database.

Each function represents a display style. Such as f1 handles math fomular, and f2 handles table. A function may alert the program that it needs a certain file in order to display correctly. For instance, to display math fomular, f1 needs MathJax.js. And program will include all needed files in the end.

For each f, it reads a string and outputs the processed string. Keywords like $fomular$ or \\begin\{tabular\} will be noticed by program and trigger the corresponding function. New function maybe call during the processing of the other function to handle Keywords nest together, such as writing math in a table. 

The design is extremely flexiable and extendable. New keywords can be added via the function with format:

```go
f(rawText string)(string, string[]) {
   do something with rawText -> rect
   needed files -> alertMessage
   return rect, alertMessage
}
```

#TODO
- [ ] MathJax
- [ ] Table
- [ ] Programming HighLight
- [ ] Figure
- [ ] Diagram
