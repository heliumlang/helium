<h1 align="center"><strong>Helium</strong></h1>
<p align="center"><em><small>Modern scripting language</small></em></p>

Helium is a simple scripting language that ensures safety, correctness and ease of use.

```rust
mod main

struct Point {
    pub int x
    pub int y
}

fn [a Point] add(Point b) Point {
    return new Point(x: a.x + b.x, y: a.y + b.y)
}

struct Person {
    pub string name
    pub string surname
    pub string fullname

    init(string name, string surname) {
        @name = name
        @surname = surname
        @fullname = name + ' ' + surname
    }
}

fn main() {
    p1, p2 := new Point(x: 5, y: 2), new Point(x: 3, y: 1)
    p3 := p1.add(p2)

    person := new Person("John", "Doe")
    println(person.fullname)
}
```
