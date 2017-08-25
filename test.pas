const
  m =  7,
  n = 85;

var
  x, y, z, q, r;

procedure multiply;
var a, b;

begin
  a := x;
  b := y;
  z := 0;
  while b > 0 do begin
    if odd b then z := z + a;
    a := 2 * a;
    b := b / 2
  end
end;

procedure divide;
var w;
begin
  r := x;
  q := 0;
  w := y;
  while w <= r do w := 2 * w;
  while w > y do begin
    q := 2 * q;
    w := w / 2;
    if w <= r then begin
      r := r - w;
      q := q + 1
    end
  end
end;

procedure gcd;
var f, g;
begin
  f := x;
  g := y;
  while f # g do begin
    if f < g then g := g - f;
    if g < f then f := f - g
  end;
  z := f
end;

begin
  read x;
  read y;
  call multiply;
  write z;
  
  read x;
  read y;
  call divide;
  write q;
  write r;
  
  read x;
  read y;
  call gcd;
  write z
end.