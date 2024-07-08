package transport

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/mkapustina/go_final_project/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

// Помощник serverError записывает сообщение об ошибке в errorLog и
// затем отправляет пользователю ответ 500 "Внутренняя ошибка сервера".
func (app *Application) ServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf(`{"error":"%s\n%s"}`, err.Error(), debug.Stack())
	app.ErrorLog.Output(2, trace)

	http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
}

// Помощник clientError отправляет определенный код состояния и
// соответствующее описание пользователю.
func (app *Application) ClientError(w http.ResponseWriter, status int, statusText string) {
	http.Error(w, fmt.Sprintf(`{"error":"%s"}`, statusText), status)
}

// Получение текущей даты без времени.
func getNowDate() time.Time {
	nowDate, _ := time.Parse("20060102", time.Now().Format("20060102"))
	return nowDate
}

// Проверка на соответствие формату полученной строики правила повтора.
func checkRepeat(repeat string) (string, []int, []int, error) {
	var num int
	var err error
	var pref string
	var intValue []string
	var suf1, suf2 []int

	if len(repeat) == 0 {
		return pref, suf1, suf2, nil
	}

	if repeat == "y" {
		pref = repeat
		return pref, suf1, suf2, nil
	}

	substrings := strings.Split(repeat, " ")
	if !(len(substrings) == 2 || len(substrings) == 3) {
		return pref, suf1, suf2, errors.New("неверный формат интервала повторения")
	}
	pref = substrings[0]

	switch pref {
	case "d":
		num, err = strconv.Atoi(substrings[1])
		if err != nil {
			return pref, suf1, suf2, errors.New("неверный формат интервала повторения")
		}
		if !(num >= 1 && num <= 400) {
			return pref, suf1, suf2, errors.New("неверное количество дней интервала повторения")
		}
		suf1 = append(suf1, num)
	case "w":
		intValue = strings.Split(substrings[1], ",")
		for i := 0; i < len(intValue); i++ {
			num, err = strconv.Atoi(intValue[i])
			if err != nil {
				return pref, suf1, suf2, errors.New("неверный формат интервала повторения")
			}
			if !(num >= 1 && num <= 7) {
				return pref, suf1, suf2, errors.New("неверный номер дня недели интервала повторения")
			}
			suf1 = append(suf1, num)
		}
	case "m":
		intValue = strings.Split(substrings[1], ",")
		for i := 0; i < len(intValue); i++ {
			num, err = strconv.Atoi(intValue[i])
			if err != nil {
				return pref, suf1, suf2, errors.New("неверный формат интервала повторения")
			}
			if !((num >= 1 && num <= 31) || num == -1 || num == -2) {
				return pref, suf1, suf2, errors.New("неверный номер дня месяца интервала повторения")
			}
			suf1 = append(suf1, num)
		}

		if len(substrings) == 3 {
			intValue = strings.Split(substrings[2], ",")
			for i := 0; i < len(intValue); i++ {
				num, err = strconv.Atoi(intValue[i])
				if err != nil {
					return pref, suf1, suf2, errors.New("неверный формат интервала повторения")
				}
				if !(num >= 1 && num <= 12) {
					return pref, suf1, suf2, errors.New("неверный номер месяца интервала повторения")
				}
				suf2 = append(suf2, num)
			}
		}
	default:
		return pref, suf1, suf2, errors.New("недопустимый тип интервала повторения")
	}
	return pref, suf1, suf2, nil
}

// Определение следующей даты повтора в соотвествии с полученными на вход
// текущей датой, датой задачи и правилами повтора
func NextDate(now time.Time, date string, repeat string) (string, error) {
	var retDate, tmpDate time.Time
	var num int
	var isValidMonth bool

	curDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", errors.New("неверный формат даты")
	}

	if len(repeat) == 0 {
		return "", errors.New("не задан интервал повторения")
	}

	pref, suf1, suf2, err := checkRepeat(repeat)
	if err != nil {
		return "", err
	}

	switch pref {
	case "y":
		retDate = curDate
		for {
			retDate = retDate.AddDate(1, 0, 0)
			if now.Before(retDate) {
				break
			}
		}
		return retDate.Format("20060102"), nil
	case "d":
		retDate = curDate
		num = suf1[0]
		for {
			retDate = retDate.AddDate(0, 0, num)
			if now.Before(retDate) {
				break
			}
		}
		return retDate.Format("20060102"), nil
	case "w":
		retDate = curDate
		if retDate.Before(now) {
			retDate = now
		}
		for {
			retDate = retDate.AddDate(0, 0, 1)
			num = int(retDate.Weekday())
			if num == 0 {
				num = 7
			}
			if slices.Contains(suf1, num) {
				break
			}
		}
		return retDate.Format("20060102"), nil
	case "m":
		retDate = curDate
		if retDate.Before(now) {
			retDate = now
		}
		for {
			retDate = retDate.AddDate(0, 0, 1)
			if len(suf2) == 0 {
				isValidMonth = true
			} else {
				isValidMonth = slices.Contains(suf2, int(retDate.Month()))
			}
			if isValidMonth {
				if slices.Contains(suf1, retDate.Day()) {
					break
				}
				// ищем 1 день след. месяца
				tmpDate = time.Date(retDate.Year(), retDate.Month(), 1, 0, 0, 0, 0, retDate.Location())
				tmpDate = tmpDate.AddDate(0, 1, 0)
				if slices.Contains(suf1, -1) {
					if retDate == tmpDate.AddDate(0, 0, -1) {
						break
					}
				}
				if slices.Contains(suf1, -2) {
					if retDate == tmpDate.AddDate(0, 0, -2) {
						break
					}
				}
			}
		}
		return retDate.Format("20060102"), nil
	default:
		return "", errors.New("недопустимый тип интервала повторения")
	}
}

// Проверка полученно даты задания на соотвествие формату
func checkDate(taskDate string, taskRepeat string) (string, error) {
	resDate := taskDate
	nowDate := getNowDate()
	if len(resDate) == 0 {
		return nowDate.Format("20060102"), nil
	}
	taskDT, err := time.Parse("20060102", resDate)
	if err != nil {
		return "", errors.New("неверный формат даты")
	}
	if taskDT.Before(nowDate) {
		if len(taskRepeat) == 0 {
			return nowDate.Format("20060102"), nil
		}
		resDate, err = NextDate(nowDate, resDate, taskRepeat)
		if err != nil {
			return "", err
		}
	}
	return resDate, nil
}

// Проверка полученного задания на соответствие формату
func CheckTask(r *http.Request) (models.Task, error) {
	var task models.Task
	var buf bytes.Buffer
	var err error

	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		return task, err
	}

	err = json.Unmarshal(buf.Bytes(), &task)
	if err != nil {
		return task, err
	}

	if len(task.Title) == 0 {
		return task, errors.New("не указан заголовок задачи")
	}

	if _, _, _, err = checkRepeat(task.Repeat); err != nil {
		return task, err
	}

	if task.Date, err = checkDate(task.Date, task.Repeat); err != nil {
		return task, err
	}

	return task, nil
}

// получение нового токена
func getToken(password string) (string, error) {
	secret := []byte("my_secret_key")

	pwdHash := sha256.Sum256([]byte(password))
	claims := jwt.MapClaims{
		"pwd": hex.EncodeToString(pwdHash[:]),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := jwtToken.SignedString(secret)
	return signedToken, err
}

// Проверка пполученной куки на соотвествие токену
func (app *Application) checkCookie(r *http.Request) error {
	if len(app.Config.Password) == 0 {
		return nil
	}

	if len(r.Header["Cookie"]) == 0 {
		return errors.New("токен отсутствует")
	}
	cookie, err := r.Cookie("token")
	if err != nil {
		return err
	}
	cookieToken := cookie.Value

	secretKey := []byte("my_secret_key")
	jwtToken, err := jwt.Parse(cookieToken, func(t *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return err
	}

	if !jwtToken.Valid {
		return errors.New("токен не валиден")
	}

	token, _ := getToken(app.Config.Password)
	if cookieToken != token {
		return errors.New("токен не валиден")
	}
	return nil
}

func (app *Application) auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := app.checkCookie(r)
		if err != nil {
			app.ClientError(w, http.StatusUnauthorized, err.Error())
		}
		next(w, r)
	})
}
