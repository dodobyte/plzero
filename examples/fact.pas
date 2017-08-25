const max = 10;
var arg;

procedure factorial;
var i, f;
begin
	i := 1;
	f := 1;
	while i <= arg do 
	begin
		f := f * i;
		i := i + 1
	end;
	write f
end;

begin
	arg := 1;
	while arg < 10 do
	begin
		call factorial;
		arg := arg + 1
	end
end.


