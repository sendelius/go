package infrastructure

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sendelius/go/apperrors"
	"gorm.io/gorm"
)

type Service[T any] struct {
	Db *gorm.DB
}

func NewService[T any](db *gorm.DB) *Service[T] {
	return &Service[T]{Db: db}
}

func (s *Service[T]) GetByID(id string) (T, error) {
	var item T
	err := s.Db.First(&item, "id = ?", id).Error
	return item, s.Error(err)
}

func (s *Service[T]) DeleteByID(id string) error {
	var item T

	tx := s.Db.Delete(&item, "id = ?", id)

	if tx.Error != nil {
		return s.Error(tx.Error)
	}

	if tx.RowsAffected == 0 {
		return apperrors.New(ErrorDelete, "запись не найдена")
	}

	return nil
}

func (s *Service[T]) CloneByID(id string) error {
	var (
		item  T
		clone T
	)

	if err := s.Db.First(&item, "id = ?", id).Error; err != nil {
		return apperrors.New(ErrorCloneNotFound, "запись не найдена")
	}

	clone = item
	value := reflect.ValueOf(&clone).Elem()
	if field := value.FieldByName("ID"); field.IsValid() && field.CanSet() {
		field.Set(reflect.Zero(field.Type()))
	}
	if field := value.FieldByName("CreatedAt"); field.IsValid() && field.CanSet() {
		field.Set(reflect.Zero(field.Type()))
	}
	if field := value.FieldByName("UpdatedAt"); field.IsValid() && field.CanSet() {
		field.Set(reflect.Zero(field.Type()))
	}

	if err := s.Db.Create(&clone).Error; err != nil {
		return s.Error(err)
	}

	return nil
}

func (s *Service[T]) DeleteBulk(ids []string) (*BulkResult, error) {
	var item T

	tx := s.Db.Where("id IN ?", ids).Delete(&item)

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
		}, apperrors.New(ErrorDeleteBulk, message)
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
		}, apperrors.New(ErrorCloneBulk, message)
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

	db := s.Db.Model(new(T))

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
	offset := (page - 1) * limit

	if page > last {
		page = last
	}

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
		return apperrors.New(ErrorUndefined, err.Error())
	}

	field := getFieldName(pgErr)

	fieldErr := func(code string, msg string) error {
		if field != "" {
			return apperrors.New(code, fmt.Sprintf("ошибка поля '%s': %s", field, msg))
		}
		return apperrors.New(code, msg)
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
		return apperrors.New(ErrorConstraintExclusion, "нарушение ограничения исключения")
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
		return apperrors.New(ErrorDivisionByZero, "деление на ноль")
	case "22004":
		return fieldErr(ErrorNullNotAllowed, "недопустимое NULL-значение")
	case "40001":
		return apperrors.New(ErrorTransactionConflict, "конфликт транзакции, повторите попытку")
	case "40P01":
		return apperrors.New(ErrorDeadlock, "взаимная блокировка транзакций, повторите попытку")
	case "42501":
		return apperrors.New(ErrorInsufficientPrivilege, "недостаточно прав")
	case "42P01":
		return apperrors.New(ErrorTableNotFound, "таблица не существует")
	case "42703":
		return fieldErr(ErrorColumnNotFound, "поле не существует")
	case "42710":
		return apperrors.New(ErrorObjectAlreadyExists, "объект уже существует")
	case "53300":
		return apperrors.New(ErrorTooManyConnections, "слишком много подключений к базе данных")
	case "55000":
		return apperrors.New(ErrorInvalidTransactionState, "операция невозможна в текущем состоянии")
	case "57014":
		return apperrors.New(ErrorQueryCanceled, "запрос был отменен")
	case "08000", "08003", "08006":
		return apperrors.New(ErrorConnection, "ошибка соединения с базой данных")
	default:
		return apperrors.New(ErrorUndefined, fmt.Sprintf("ошибка базы данных (%s): %s", pgErr.Code, pgErr.Error())) // TODO: убрать pgErr.Error() в какой то лог, а на фронте пускай будет просто код ошибки
	}
}
