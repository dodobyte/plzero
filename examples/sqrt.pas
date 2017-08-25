var x, s;

procedure sqrt;
var i;
begin
	x := x * 4;
	s := 1000;
	i := 0;
	while i < 10 do
	begin
		s := s - (s * s - x) / (2 * s);
		i := i + 1
	end;
	s := s / 2
end;

begin
	while 1 = 1 do
	begin
		read x;
		call sqrt;
		write s
	end
end.