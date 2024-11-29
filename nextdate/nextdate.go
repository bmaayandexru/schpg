package nextdate

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// типы операций
const (
	operationTypes string = "dywm"
	template              = "20060102"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	startDate, err := time.Parse(template, date)
	if err != nil {
		return "", fmt.Errorf("ошибка в стартовой дате %v\n", err)
	}
	repSlice := strings.Split(repeat, " ") // разложить repeat в слайс строк
	if len(repSlice[0]) == 0 {
		return "", errors.New("правило повторения не задано")
	}
	// тут слайс не пустой. проверяем певый элемент на соответствие
	// первая строка - тип повторения. один символ из стоки "dywm"
	if len(repSlice[0]) != 1 {
		return "", errors.New("Длина типа не равна 1")
	}
	if !strings.Contains(operationTypes, repSlice[0]) {
		return "", errors.New("Неизвестный тип операции повторения")
	}
	switch repSlice[0] {
	case "d": // d дни
		return NextDay(now, startDate, repSlice)
	case "y": // y год
		return NextYear(now, startDate, repSlice)
	case "w": // w дни недели
		return NextWeekDay(now, startDate, repSlice)
	case "m": // m дни месяца
		return NextDateMonth(now, startDate, repSlice)
	} // switch
	return "", errors.New("не удалось определить следующую дату")
}

func makeSlice3Month(date time.Time) []time.Time {
	retSl := make([]time.Time, 0, 3)
	retSl = append(retSl, time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local))
	date = date.AddDate(0, 1, 0)
	retSl = append(retSl, time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local))
	date = date.AddDate(0, 1, 0)
	retSl = append(retSl, time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local))
	return retSl
}

func checkDay(year int, month int, day int) bool {
	// месяцы в которых нет 31го числа
	msless31 := map[int]bool{2: true, 4: true, 6: true, 9: true, 11: true}

	if (day >= 1 && day <= 28) || day == -1 || day == -2 {
		// дни [1..28] проходят проверку
		return true
	} else {
		if day == 0 {
			return false
		}
		// проверка наличия в месяце 29, 30, 31 чисел
		if day == 31 && msless31[month] {
			// 31е число есть не во всех месяцах
			return false
		}
		if year%4 == 0 {
			// високосный год
			if month == 2 && day == 30 {
				return false
			}
		} else {
			// не високосный год
			if month == 2 && (day == 29 || day == 30) {
				return false
			}
		}
		// проверка прошла возвращаем true
		return true
	}
}

func NextDateMonth(now time.Time, startDate time.Time, repSlice []string) (string, error) {
	if len(repSlice) < 2 {
		return "", errors.New("m: не узазана дата/даты месяца")
	}
	// проверка списка первого параметра [1..31,-1,-2]
	repSlice1 := strings.Split(repSlice[1], ",") // из 2й группы делаем слас строк
	sliDays := make([]int, 0, len(repSlice1))    // создаем слайс дней
	for _, strDay := range repSlice1 {           // []string --> []int
		iDay, err := strconv.Atoi(strDay)
		if err == nil {
			if ((iDay >= 1) && (iDay <= 31)) || iDay == -1 || iDay == -2 { // если в диапазане
				sliDays = append(sliDays, iDay)
			} else {
				return "", errors.New("m: день месяца вне диапазона") // ошибка
			}
		} else {
			return "", fmt.Errorf("m: День указан не числом. Ошибка:%v \n", err)
		}
	}
	sort.Ints(sliDays)        // в slDays отсортированные дни
	var sldMonths []time.Time // слайс первых чисел нужных месяцев
	if len(repSlice) == 2 {
		// тут только один список. создаем слайс из трёх месяцев
		if startDate.Before(now) {
			sldMonths = makeSlice3Month(now)
		} else {
			sldMonths = makeSlice3Month(startDate)
		}
	} else {
		// тут два списка len >= 3
		// проверяем второй список на корректность [1..12]
		repSlice2 := strings.Split(repSlice[2], ",") // третий параметр в слайс строк
		sliMonths := make([]int, 0, len(repSlice2))  // слайс целых чисел месяцев
		for _, strMonth := range repSlice2 {
			iMonth, err := strconv.Atoi(strMonth)
			if err == nil {
				if (iMonth >= 1) && (iMonth <= 12) { // если в диапазоне
					sliMonths = append(sliMonths, iMonth) // добавляем
				} else {
					return "", errors.New("m: месяц за пределами диапазона") // ошибка
				}
			} else {
				return "", errors.New("m: указано не число") // в слайсе не число
			}
		}
		sort.Ints(sliMonths) // сортировка
		// создаем слайс дат для текущего и следующего года
		sldYears := make([]time.Time, 0)
		sldYears = append(sldYears, now)
		sldYears = append(sldYears, now.AddDate(1, 0, 0))
		// создаем слайс из дат для каждого года
		for _, dYear := range sldYears {
			for _, iMonth := range sliMonths {
				// добавляем в slMonths дату (dYear,iMonth,01)
				sldMonths = append(sldMonths, time.Date(dYear.Year(), time.Month(iMonth), 1, 0, 0, 0, 0, time.Local))
			}
		}
	}
	// тут сфоормированы слайсы из дат для месяцев и годов
	// формируем слайс строк дат по дням, т к требуется сортировка
	slsDays := make([]string, 0)
	for _, dMonth := range sldMonths {
		for _, iDay := range sliDays {
			// создаем список конкретных дат
			// из даты dMonth берем месяц и год
			// проверяем есть ли iDay в этом месяце
			if checkDay(dMonth.Year(), int(dMonth.Month()), iDay) {
				// если есть формируем дату dMonth.Year(), dMonth.Month(). iDay
				// добавляем строкой в список дней, который потом отсортируем и выберем нужный
				if iDay < 0 {
					dM := dMonth.AddDate(0, 1, iDay)
					slsDays = append(slsDays, dM.Format(template))
				} else {
					slsDays = append(slsDays, time.Date(dMonth.Year(), dMonth.Month(), iDay, 0, 0, 0, 0, time.Local).Format(template))
				}
			}
		}
	}
	// сортируем список из строк дат
	if len(slsDays) == 0 {
		return "", errors.New("m: Не возможно определить дату по указанным параметрам (m 30,31 2)")
	}
	sort.Strings(slsDays)
	// выбираем нужный день и возвращаем
	for _, sDay := range slsDays {
		dDay, _ := time.Parse(template, sDay)
		if now.Before(dDay) {
			return sDay, nil
		}
	}
	// если список пуст то это m 30,31 2
	return "", errors.New("m: Не возможно определить дату по указанным параметрам (m 30,31 2)")
}

func NextDay(now time.Time, startDate time.Time, repSlice []string) (string, error) {
	if len(repSlice) < 2 {
		return "", errors.New("d: нет указания дней")
	}
	if len(repSlice) > 2 {
		return "", errors.New("d: много параметров")
	}
	// разложить rs[1] на слайс
	repSlice1 := strings.Split(repSlice[1], ",")
	if len(repSlice1) != 1 {
		return "", errors.New("d: число дней указано не одним числом")
	}
	dcount, err := strconv.Atoi(repSlice1[0])
	if err != nil {
		return "", errors.New("d: параметр не число") // параметр не число
	}
	// число от 1 до 400 включительно
	if (dcount < 1) || (dcount > 400) {
		return "", errors.New("d: число вне диапазона (<1 >400)") // число вне диапазона
	}
	// тут всё корректно. можно возвращать значение
	for {
		startDate = startDate.AddDate(0, 0, dcount)
		if startDate.Format(template) >= now.Format(template) {
			break
		}
	}
	return startDate.Format(template), nil
}

func NextYear(now time.Time, startDate time.Time, repSlice []string) (string, error) {
	// !!! в любом случае идет перенос даты на год хотя бы однократно.
	if len(repSlice) != 1 {
		return "", errors.New("y: количество параметров != 0") // ошибка количества параметров
	}
	for {
		startDate = startDate.AddDate(1, 0, 0)
		if startDate.After(now) {
			break
		}
	}
	return startDate.Format(template), nil
}

func NextWeekDay(now time.Time, startDate time.Time, repSlice []string) (string, error) {
	// w дни недели
	// у w может быть слайс из 1-7
	if len(repSlice) != 2 {
		return "", errors.New("w: количество параметров != 2") // ошибка количества параметров
	}
	repSlice1 := strings.Split(repSlice[1], ",") // второй параметр в слайс строк
	for i := 0; i < len(repSlice1); i++ {        // проверка на допустимые значения
		if (repSlice1[i] < "1") || (repSlice1[i] > "7") { // если вне диапазона
			return "", errors.New("w: один из параметров за пределами диапазона (<1 >7)") // ошибка
		}
	}
	// мапа из значений с преобразование в int'ы а затем в Weedday
	mapWeekDays := make(map[time.Weekday]bool)
	for _, strDay := range repSlice1 {
		iDay, _ := strconv.Atoi(strDay)
		if iDay == 7 {
			iDay = 0 // вс теперь 0 а не 7
		}
		mapWeekDays[time.Weekday(iDay)] = true
	}
	curDay := now
	if startDate.After(now) {
		curDay = startDate
	}
	for {
		curDay = curDay.AddDate(0, 0, 1)
		_, found := mapWeekDays[curDay.Weekday()]
		if found {
			break
		}
	}
	return curDay.Format(template), nil
}
