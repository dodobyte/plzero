const max = 10;

procedure fibonacci;
var i, p, q, t;
begin
	p := 0;
	q := 1;
	i := 0;
	while i < max do
	begin
		write p + q;
		t := p;
		p := p + q;
		q := t;
		i := i + 1
	end
end;

call fibonacci
.