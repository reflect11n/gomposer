package interfaces

type Window interface {
	Render()
	HandleEvent()
	Modify() //изменение размеров, isActive
	Translate() //если мы разрешаем доступ для трансляции
}
