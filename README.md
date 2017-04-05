# treewrite

`treewrite` is a tool for performing automated replacements in text
files with nested structured.  In particular, it is useful for making
changes to programs written in C/C++.

## Examples

Here are some sample command lines.  Keep reading to understand what these
commands do:

```shell
treewrite `printf(` `fprintf(stdout,`
treewrite `fprintf(stdout, $args*)` `printf($args*)`
treewrite 'bcopy($src, $dst, $size)' 'memcpy($dst, $src, $size)'
```

## Motivation

For simple replacements people can use tools like `sed` and `perl`.
However, those tools have several shortcomings, mainly due to their
lack of awareness of the underlying syntax.  This is best illustrated
with an example.  Suppose we want to replace the use of `bcopy` with
`memset`.  As an example, the following code:

```none
bcopy(&src, &dst, sizeof(dst))
```

should be converted to:

```none
memcpy(&dst, &src, sizeof(dst))
```

With regular-expression based replacement as supported by `perl` or
`sed`, we might write a substitution command that looks something like:

```
s/bcopy\(([^,]+), ([^,]+), ([^,]+)\)/memset($2, $1, $3)/
```

The preceding command is very error-prone.  E.g., it breaks if spacing
is not quite as expected, if there are comments present inside the
call, if one of the arguments is more complicated and contains a
comma, etc.

With `treewrite`, instead of using regular expressions to match the
string form of a tree, the pattern is parsed into a tree and matched
against the tree representation of the program.  The preceding replacement
can be achieved by the following command:

```shell
treewrite 'bcopy($src, $dst, $size)' 'memcpy($dst, $src, $size)'
```

The `treewrite` command will work even if the text to be matched
contains comments or is formatted unexpectedly (e.g., if it is split
into multiple lines).  It will also deal properly with complicated
arguments to `bcopy` since the variables `$src`, `$dst`, `$size` will
match entire sub-expressions.

## Running as a Filter

A common way to use `treewrite` is as a filter that reads from
standard input, finds occurrences of a supplied pattern, replaces each
occurrence with a supplied replacement, and prints the result to
standard output.  E.g., to replace `bcopy` calls with `memcpy` and
to examine the changes made, the following command suffices (the
pattern is supplied as the first argument, and the replacement
as the second argument):

```shell
cat srcfile | \
treewrite 'bcopy($src, $dst, $size)' 'memcpy($dst, $src, $size)' | \
diff srcfile -
```

## In-place editing of files

If the `-edit` flag is specified, `treewrite` will read from each
file, replace pattern with replacement, and write the result back to
each file.  So to apply the preceding change to all C files in a
directory:

```shell
treewrite -edit 'bcopy($src, $dst, $size)' 'memcpy($dst, $src, $size)' *.c
```

## Matching Process

The input text and the pattern are both parsed into trees according to
the input language syntax.  `treewrite` then finds all sub-trees in
the input tree that match the pattern tree.  Normal text in the
pattern must match input text exactly.  White-space and comments are
ignored.  The matched input sub-tree is replaced by the replacement.

## Variable Matching and Substitution

As described above, a pattern can contain variables.  E.g., `$src`,
`$dst`, `$size` in the `bcopy/memcpy` example.  Variables start with a
`$` sign followed by an identifier.  When the pattern matches, each
corresponding portion node in the matched input tree is assigned to
the variable.  So if we applied the preceding `bcopy` pattern to the
text `bcopy(f(x,y), dst, n)`, `$src` will be assigned the value `f(x,y)`.

Variables can be referenced in the replacement text.  When generating
the replacement each variable occurrence is replaced with the value
assigned to that variable.

## Repeated Variables

A normal variable matches a single expression/node in the input tree.
Sometimes we want to match a whole sequence of expressions.  E.g., if
we want to replace `fprintf(stdout, ...)` with direct calls to
`printf`, we would want to capture any number of expressions following
`stdout` into a variable.  A `repeating variable` can match zero or more
expressions.  A repeating variable uses the syntax `$<identifier>*`.  So
the `printf` replacement can be done by:

```shell
treewrite `fprintf(stdout, $args*)` `printf($args*)`
```
If the input text contains `fprintf(stdout, "%s:%d", host, port)`, `$args*`
will end up matching `"%s:%d", host, port`.

## Reading Pattern and Replacement from a File

As patterns and replacements get more complicated, it becomes unwieldy
to specify them on the command line.  Instead, they can be placed in a
file (with a line of at least three dashes separating the pattern from
the replacement) and the file name supplied using the `-apply` flag to
`treewrite`.  For example, place the following in a file named
`replacement`:

```none
fprintf(stdout, $args*)
----
printf($args*)
```

and then run:


```shell
treewrite -apply replacement *.c
```

## Caveats

*   Parsing C and C++ is hard. This tool implements heuristic based
    parsing which can be easily thrown off, e.g., by a confusion between
    the use of `<` and `>` in C++ templates as opposed to expressions.
*   Comments in replaced text are preserved, but may end up in an
    unexpected order.
*   Multi-line replacements are unlikely to be indented/formatted
    properly. After running `treewrite`, users should run a tool like
    `clang-format` to fix up formatting.
