package infrastructure

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sendelius/go/apperrors"
	"github.com/sendelius/go/database"
	"gorm.io/gorm"
)

type Service[T any] struct {
	db *database.Service
}

func NewService[T any](db *database.Service) *Service[T] {
	return &Service[T]{db: db}
}

func (s *Service[T]) Db() *gorm.DB {
	return s.db.Db()
}

func (s *Service[T]) GetByID(id string) (T, error) {
	var item T
	err := s.Db().First(&item, "id = ?", id).Error
	return item, s.Error(err)
}

func (s *Service[T]) DeleteByID(id string) error {
	var item T

	tx := s.Db().Delete(&item, "id = ?", id)

	if tx.Error != nil {
		return s.Error(tx.Error)
	}

	if tx.RowsAffected == 0 {
		return apperrors.New(ErrorDelete, "запись не найдена").Build()
	}

	return nil
}

func (s *Service[T]) CloneByID(id string) error {
	var item T

	if err := s.Db().First(&item, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(ErrorCloneNotFound, "запись не найдена").Cause(err).Build()
		}
		return s.Error(err)
	}

	value := reflect.ValueOf(item)
	var clone any

	switch value.Kind() {
	case reflect.Pointer:
		if value.IsNil() || value.Elem().Kind() != reflect.Struct {
			return apperrors.New(ErrorCloneTypes, "тип записи для клонирования должен быть структурой").Build()
		}

		cloneValue := reflect.New(value.Elem().Type())
		cloneValue.Elem().Set(value.Elem())
		resetCloneFields(cloneValue.Elem())
		clone = cloneValue.Interface()

	case reflect.Struct:
		cloneValue := reflect.New(value.Type())
		cloneValue.Elem().Set(value)
		resetCloneFields(cloneValue.Elem())
		clone = cloneValue.Interface()

	default:
		return apperrors.New(ErrorCloneTypes, "тип записи для клонирования должен быть структурой").Build()
	}

	if err := s.Db().Create(clone).Error; err != nil {
		return s.Error(err)
	}

	return nil
}

func (s *Service[T]) DeleteBulk(ids []string) (*BulkResult, error) {
	var item T

	tx := s.Db().Where("id IN ?", ids).Delete(&item)

	if tx.Error != nil {
		return &BulkResult{
			Success: 0,
			Failed:  int64(len(ids)),
		}, s.Error(tx.Error)
	}

	failed := int64(len(ids)) - tx.RowsAffected
	if failed > 0 {
		message := "не удалось удалить записи"
		if tx.RowsAffected > 0 {
			message = "некоторые записи не удалось удалить"
		}
		return &BulkResult{
			Success: tx.RowsAffected,
			Failed:  failed,
		}, apperrors.New(ErrorDeleteBulk, message).Build()
	}

	return &BulkResult{
		Success: tx.RowsAffected,
		Failed:  failed,
	}, nil
}

func (s *Service[T]) CloneBulk(ids []string) (*BulkResult, error) {
	var (
		success int64
		failed  int64
	)

	for _, id := range ids {
		if err := s.CloneByID(id); err != nil {
			failed++
			continue
		}
		success++
	}

	if failed > 0 {
		message := "не удалось клонировать записи"
		if success > 0 {
			message = "некоторые записи не удалось клонировать"
		}

		return &BulkResult{
			Success: success,
			Failed:  failed,
		}, apperrors.New(ErrorCloneBulk, message).Build()
	}

	return &BulkResult{
		Success: success,
		Failed:  failed,
	}, nil
}

func (s *Service[T]) GetListRequest() *ListRequest {
	return &ListRequest{}
}

func (s *Service[T]) GetList(request *ListRequest, conf *ListConfig) ([]T, int64, ListPagination, error) {
	if conf == nil {
		conf = defaultListConfig()
	}

	var (
		items      []T
		total      int64
		pagination ListPagination
	)

	db := s.Db().Model(new(T))

	// фильтры
	// TODO: доработать фильтры по типу
	for key, value := range request.Filters {
		allowed := false
		for _, field := range conf.FilterFields {
			if key == field {
				allowed = true
				break
			}
		}
		if allowed {
			db = db.Where(key+" = ?", value)
		}
	}

	// поиск
	if request.Search.Value != "" {

		searchFields := conf.SearchFields

		if len(searchFields) == 0 {
			searchFields = []string{"title", "slug"}
		}

		// если поле поиска указано явно
		if request.Search.Key != "" {
			requestFields := strings.Split(
				request.Search.Key,
				",",
			)

			var allowed []string
			for _, field := range requestFields {
				field = strings.TrimSpace(field)

				for _, available := range searchFields {
					if field == available {
						allowed = append(allowed, field)
					}
				}
			}

			if len(allowed) > 0 {
				searchFields = allowed
			}
		} else {
			// если поле не указано — берем первое
			searchFields = searchFields[:1]
		}

		var conditions []string
		var args []any
		for _, field := range searchFields {
			if conf.SearchSimilarity {
				conditions = append(conditions, field+" % ?")
			} else {
				conditions = append(conditions, field+" ILIKE ?")
			}
			args = append(args, request.Search.Value)
		}

		if len(conditions) > 0 {
			db = db.Where(strings.Join(conditions, " OR "), args...)

			if conf.SearchSimilarity {
				value := strings.ReplaceAll(request.Search.Value, "'", "''")
				var orderSimilarity []string
				for _, field := range searchFields {
					orderSimilarity = append(orderSimilarity, field+" = '"+value+"' DESC", "similarity("+field+", '"+value+"') DESC")
				}
				db = db.Order(strings.Join(orderSimilarity, ", "))
			}
		}
	}

	// получаем общее кол-во записей
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, pagination, s.Error(err)
	}

	// сортировка
	if request.Sort.Key != "" {
		allowed := false
		for _, field := range conf.SortFields {
			if request.Sort.Key == field {
				allowed = true
				break
			}
		}
		if allowed {
			order := "ASC"
			if request.Sort.Type == "desc" {
				order = "DESC"
			}
			db = db.Order(
				request.Sort.Key + " " + order,
			)
		}
	}

	// пагинация
	page := request.Page
	if page <= 0 {
		page = 1
	}

	limit := request.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	last := int(math.Ceil(float64(total) / float64(limit)))
	if last == 0 {
		page = 1
	} else if page > last {
		page = last
	}

	offset := (page - 1) * limit
	if page < last {
		pagination.Next = new(page + 1)
	}
	if page > 1 {
		pagination.Prev = new(page - 1)
	}

	// получение записей
	db = db.Limit(limit).Offset(offset)
	if err := db.Find(&items).Error; err != nil {
		return nil, 0, pagination, s.Error(err)
	}

	pagination.Page = page
	pagination.Limit = limit
	pagination.Last = last

	return items, total, pagination, nil
}

func (s *Service[T]) Error(err error) error {
	if err == nil {
		return nil
	}

	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if !ok {
		return apperrors.New(ErrorUndefined, "внутренняя ошибка базы данных").Cause(err).Build()
	}

	field := getFieldName(pgErr)

	fieldErr := func(code string, msg string) error {
		if field != "" {
			return apperrors.New(code, fmt.Sprintf("ошибка поля '%s': %s", field, msg)).Build()
		}
		return apperrors.New(code, msg).Build()
	}

	switch pgErr.Code {
	case "23502":
		return fieldErr(ErrorConstraintNotNull, "обязательное поле")
	case "23503":
		return fieldErr(ErrorConstraintForeignKey, "связанный объект не найден")
	case "23505":
		return fieldErr(ErrorConstraintUnique, "значение уже существует")
	case "23514":
		return fieldErr(ErrorConstraintCheck, "значение не проходит проверку")
	case "23P01":
		return apperrors.New(ErrorConstraintExclusion, "нарушение ограничения исключения").Build()
	case "22001":
		return fieldErr(ErrorValueTooLong, "слишком длинное значение")
	case "22003":
		return fieldErr(ErrorNumericOutOfRange, "число выходит за допустимый диапазон")
	case "22P02":
		return fieldErr(ErrorInvalidTextFormat, "неверный формат значения")
	case "22007":
		return fieldErr(ErrorInvalidDatetimeFormat, "неверный формат даты или времени")
	case "22008":
		return fieldErr(ErrorDatetimeOutOfRange, "некорректная дата или время")
	case "22012":
		return apperrors.New(ErrorDivisionByZero, "деление на ноль").Build()
	case "22004":
		return fieldErr(ErrorNullNotAllowed, "недопустимое NULL-значение")
	case "40001":
		return apperrors.New(ErrorTransactionConflict, "конфликт транзакции, повторите попытку").Build()
	case "40P01":
		return apperrors.New(ErrorDeadlock, "взаимная блокировка транзакций, повторите попытку").Build()
	case "42501":
		return apperrors.New(ErrorInsufficientPrivilege, "недостаточно прав").Build()
	case "42P01":
		return apperrors.New(ErrorTableNotFound, "таблица не существует").Build()
	case "42703":
		return fieldErr(ErrorColumnNotFound, "поле не существует")
	case "42710":
		return apperrors.New(ErrorObjectAlreadyExists, "объект уже существует").Build()
	case "53300":
		return apperrors.New(ErrorTooManyConnections, "слишком много подключений к базе данных").Build()
	case "55000":
		return apperrors.New(ErrorInvalidTransactionState, "операция невозможна в текущем состоянии").Build()
	case "57014":
		return apperrors.New(ErrorQueryCanceled, "запрос был отменен").Build()
	case "08000", "08003", "08006":
		return apperrors.New(ErrorConnection, "ошибка соединения с базой данных").Build()
	default:
		return apperrors.New(ErrorUndefined, "внутренняя ошибка базы данных").Cause(pgErr).Build()
	}
}
