# Expression

## Operator Precedence
| Precedence | Operator type                    | Associativity | Individual operators | NOTES |
| ---------- | -------------------------------- | ------------- | -------------------- | ----- |
| 18         | Grouping                         | n/a           | ( … )                |       |
| 17         | Member Access                    | left-to-right | … . …                |       |
| 17         | Computed Member Access           | n/a           | … [ … ]              |       |
| 17         | Function Call                    | n/a           | … ( … )              |       |
| 15         | Postfix Increment                | n/a           | … ++                 |       |
| 15         | Postfix Decrement                | n/a           | … --                 |       |
| 14         | Logical NOT (!)                  | n/a           | ! …                  |       |
| 14         | Bitwise NOT (~)                  | n/a           | ~ …                  | TBD   |
| 14         | Unary plus (+)                   | n/a           | + …                  |       |
| 14         | Unary negation (-)               | n/a           | - …                  |       |
| 14         | Prefix Increment                 | n/a           | ++ …                 |       |
| 14         | Prefix Decrement                 | n/a           | -- …                 |       |
| 13         | Exponentiation (\*\*)            | right-to-left | … \*\* …             |       |
| 12         | Multiplication (\*)              | left-to-right | … \* …               |       |
| 12         | Division (/)                     | left-to-right | … / …                |       |
| 12         | Remainder (%)                    | left-to-right | … % …                |       |
| 11         | Addition (+)                     | left-to-right | … + …                |       |
| 11         | Subtraction (-)                  | left-to-right | … - …                |       |
| 10         | Bitwise Left Shift (<<)          | left-to-right | … << …               | TBD   |
| 10         | Bitwise Right Shift (>>)         | left-to-right | … >> …               | TBD   |
| 9          | Less Than (<)                    | left-to-right | … < …                |       |
| 9          | Less Than Or Equal (<=)          | left-to-right | … <= …               |       |
| 9          | Greater Than (>)                 | left-to-right | … > …                |       |
| 9          | Greater Than Or Equal (>=)       | left-to-right | … >= …               |       |
| 9          | in                               | left-to-right | … in …               |       |
| 8          | Equality (==)                    | left-to-right | … == …               |       |
| 8          | Inequality (!=)                  | left-to-right | … != …               |       |
| 7          | Bitwise AND (&)                  | left-to-right | … & …                | TBD   |
| 6          | Bitwise XOR (^)                  | left-to-right | … ^ …                | TBD   |
| 5          | Bitwise OR ( \| )                | left-to-right | … \| …               | TBD   |
| 4          | Logical AND (&&)                 | left-to-right | … && …               |       |
| 3          | Logical OR ( \|\| )              | left-to-right | … \|\| …             |       |
| 3          | Nullish coalescing operator (??) | left-to-right | … ?? …               |       |
| 2          | Conditional (ternary) operator   | right-to-left | … ? … : …            |       |
| 1          | Comma                            | left-to-right | … , …                |       |

### Conditional (ternary)  Operator 
If a condition followed by a question mark (?), then an expression to execute if the condition is [truthy](#truthy) followed by a colon (:), and finally the expression to execute if the condition is [falsy](#falsy). This operator is frequently used as an alternative to an if...else statement.

```
age := 26
age >= 21 ? "Beer" : "Juice" // "Beer" 
```

### Nullish coalescing operator (??)

The nullish coalescing operator (??) is a logical operator that returns its right-hand side operand when its left-hand side operand is null or undefined, and otherwise returns its left-hand side operand.

This can be seen as a special case of the [logical OR (||) operator](#logical-or-), which returns the right-hand side operand if the left operand is any [falsy](#falsy) value, not only null or undefined. In other words, if you use || to provide some default value to another variable foo, you may encounter unexpected behaviors if you consider some falsy values as usable (e.g., '' or 0). See below for more examples.

In this example, we will provide default values but keep values other than null or undefined.
```
nullValue := null;
emptyText := ""; // falsy
someNumber := 42;

valA = nullValue ?? "default for A"; //default for A
valB = emptyText ?? "default for B"; //""
valC = someNumber ?? 0;              //42
```

### Logical OR (||)
The logical OR (||) operator (logical disjunction) for a set of operands is true if and only if one or more of its operands is true. It is typically used with boolean (logical) values. When it is, it returns a Boolean value. 

#### non-Boolean value
However, the || operator actually returns the value of one of the specified operands, so if this operator is used with non-Boolean values, it will return a non-Boolean value.
```
expr1 || expr2
```
If expr1 can be converted to true(so-called  [truthy](#truthy)), returns expr1; else, returns expr2.


```
true  || true       // t || t returns true
false || true       // f || t returns true
true  || false      // t || f returns true
false || (3 === 4)  // f || f returns false
'Cat' || 'Dog'      // t || t returns "Cat"
false || 'Cat'      // f || t returns "Cat"
'Cat' || false      // t || f returns "Cat"
''    || false      // f || f returns false
false || ''         // f || f returns ""
false || varObject // f || object returns varObject
```

### Logical AND (&&)

The logical AND (&&) operator (logical conjunction) for a set of boolean operands will be true if and only if all the operands are true. Otherwise it will be false.

#### non-Boolean value
The operator returns the value of the first [falsy](#falsy) operand encountered when evaluating from left to right, or the value of the last operand if they are all truthy.

```
true && true // t && t returns true
true && false // t && f returns false
false && true // f && t returns false
false && (3 === 4) // f && f returns false
'Cat' && 'Dog' // t && t returns "Dog"
false && 'Cat' // f && t returns false
'Cat' && false // t && f returns false
'' && false // f && f returns ""
false && '' // f && f returns false
```
#### Truthy
A truthy value is a value that is considered true when encountered in a Boolean context. All values are truthy unless they are defined as falsy. That is, all values are truthy except false, 0, "", null, nil, and NaN.


#### Falsy

Examples of expressions that can be converted to false are:

```
nil
NaN
0
empty string ("" or '' or ``)
null
```