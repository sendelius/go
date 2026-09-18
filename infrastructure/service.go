package infrastructure

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type BaseService[T any] struct {
	Db *gorm.DB
}

func NewBaseService[T any](db *gorm.DB) *BaseService[T] {
	return &BaseService[T]{Db: db}
}

func (s *BaseService[T]) GetByID(id string) (T, error) {
	var item T
	err := s.Db.First(&item, "id = ?", id).Error
	return item, s.Error(err)
}

type GetListConfig struct {
	SearchFields     []string
	SearchSimilarity bool
	FilterFields     []string
	SortFields       []string
}

func DefaultGetListConfig() *GetListConfig {
	return &GetListConfig{
		SearchFields: []string{
			"title",
			"slug",
		},
		SearchSimilarity: false,
		FilterFields:     []string{},
		SortFields: []string{
			"id",
		},
	}
}

func (s *BaseService[T]) GetList(filter *BaseRequestGetList, conf *GetListConfig) ([]T, int64, BasePagination, error) {
	if conf == nil {
		conf = DefaultGetListConfig()
	}

	var (
		items      []T
		total      int64
		pagination BasePagination
	)

	db := s.Db.Model(new(T))

	// фильтры
	// TODO: доработать фильтры по типу
	for key, value := range filter.Filters {
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
	if filter.Search.Value != "" {

		searchFields := conf.SearchFields

		if len(searchFields) == 0 {
			searchFields = []string{"title", "slug"}
		}

		// если поле поиска указано явно
		if filter.Search.Key != "" {
			requestFields := strings.Split(
				filter.Search.Key,
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
			args = append(args, filter.Search.Value)
		}

		if len(conditions) > 0 {
			db = db.Where(strings.Join(conditions, " OR "), args...)

			if conf.SearchSimilarity {
				value := strings.ReplaceAll(filter.Search.Value, "'", "''")
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
	if filter.Sort.Key != "" {
		allowed := false
		for _, field := range conf.SortFields {
			if filter.Sort.Key == field {
				allowed = true
				break
			}
		}
		if allowed {
			order := "ASC"
			if filter.Sort.Type == "desc" {
				order = "DESC"
			}
			db = db.Order(
				filter.Sort.Key + " " + order,
			)
		}
	}

	// пагинация
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	limit := filter.Limit
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

func (s *BaseService[T]) DeleteByID(id string) error {
	var item T

	tx := s.Db.Delete(&item, "id = ?", id)

	if tx.Error != nil {
		return s.Error(tx.Error)
	}

	if tx.RowsAffected == 0 {
		return errors.New("запись не найдена")
	}

	return nil
}

type BulkResult struct {
	Success int64 `json:"success"`
	Failed  int64 `json:"failed"`
}

func (s *BaseService[T]) DeleteBulk(ids []string) (*BulkResult, error) {
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
		}, errors.New(message)
	}

	return &BulkResult{
		Success: tx.RowsAffected,
		Failed:  failed,
	}, nil
}

func (s *BaseService[T]) CloneByID(id string) error {
	var (
		item  T
		clone T
	)

	if err := s.Db.First(&item, "id = ?", id).Error; err != nil {
		return errors.New("запись не найдена")
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

func (s *BaseService[T]) CloneBulk(ids []string) (*BulkResult, error) {
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
		}, errors.New(message)
	}

	return &BulkResult{
		Success: success,
		Failed:  failed,
	}, nil
}

/* --------------------------------------------- REQUEST --------------------------------------------- */

type BaseRequestGetList struct {
	Limit   int                      `json:"limit" validate:"omitempty,min=1,max=1000"`
	Page    int                      `json:"page" validate:"omitempty,min=1"`
	Sort    BaseRequestGetListSort   `json:"sort" validate:"omitempty"`
	Search  BaseRequestGetListSearch `json:"search" validate:"omitempty"`
	Filters map[string]string        `json:"filters" validate:"omitempty"`
}

type BaseRequestGetListSort struct {
	Key  string `json:"key"`
	Type string `json:"type" validate:"omitempty,oneof=asc desc"`
}

type BaseRequestGetListSearch struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (r *BaseRequestGetList) SetQueryParam(key string, value any) {
	switch key {
	case "filters":
		if filters, ok := value.(map[string]string); ok {
			r.Filters = filters
		}
	case "search":
		if search, ok := value.(map[string]string); ok {
			r.Search.Key = search["key"]
			r.Search.Value = search["value"]
		}
	case "sort":
		if sort, ok := value.(map[string]string); ok {
			r.Sort.Key = sort["key"]
			r.Sort.Type = sort["type"]
		}
	}
}

func (s *BaseService[T]) RequestGetList() *BaseRequestGetList {
	return &BaseRequestGetList{}
}

/* -------------------------------------------- RESPONSE -------------------------------------------- */

type BasePagination struct {
	Page  int  `json:"page"`
	Limit int  `json:"limit"`
	Last  int  `json:"last"`
	Next  *int `json:"next,omitempty"`
	Prev  *int `json:"prev,omitempty"`
}

/* -------------------------------------------- ERRORS -------------------------------------------- */

func (s *BaseService[T]) Error(err error) error {
	if err == nil {
		return nil
	}

	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if !ok {
		return err
	}

	field := getFieldName(pgErr)

	fieldErr := func(msg string) error {
		if field != "" {
			return fmt.Errorf("ошибка поля '%s': %s", field, msg)
		}
		return errors.New(msg)
	}

	switch pgErr.Code {

	// integrity constraint violation
	case "23502":
		return fieldErr("обязательное поле")

	case "23503":
		return fieldErr("связанный объект не найден")

	case "23505":
		return fieldErr("значение уже существует")

	case "23514":
		return fieldErr("значение не проходит проверку")

	case "23P01":
		return errors.New("нарушение ограничения исключения")

	// string / data
	case "22001":
		return fieldErr("слишком длинное значение")

	case "22003":
		return fieldErr("число выходит за допустимый диапазон")

	case "22P02":
		return fieldErr("неверный формат значения")

	case "22007":
		return fieldErr("неверный формат даты или времени")

	case "22008":
		return fieldErr("некорректная дата или время")

	case "22012":
		return errors.New("деление на ноль")

	case "22004":
		return fieldErr("недопустимое NULL-значение")

	// transaction
	case "40001":
		return errors.New("конфликт транзакции, повторите попытку")

	case "40P01":
		return errors.New("взаимная блокировка транзакций, повторите попытку")

	// permissions
	case "42501":
		return errors.New("недостаточно прав")

	// objects
	case "42P01":
		return errors.New("таблица не существует")

	case "42703":
		return fieldErr("поле не существует")

	case "42710":
		return errors.New("объект уже существует")

	// resources
	case "53300":
		return errors.New("слишком много подключений к базе данных")

	case "55000":
		return errors.New("операция невозможна в текущем состоянии")

	case "57014":
		return errors.New("запрос был отменен")

	// connection
	case "08000", "08003", "08006":
		return errors.New("ошибка соединения с базой данных")

	default:
		return fmt.Errorf("ошибка базы данных (%s): %s", pgErr.Code, pgErr.Error()) // TODO: убрать pgErr.Error() в какой то лог, а на фронте пускай будет просто код ошибки
	}
}

var keyRe = regexp.MustCompile(`Key \((.*?)\)=`)

func getFieldName(pgErr *pgconn.PgError) string {
	if pgErr.ColumnName != "" {
		return pgErr.ColumnName
	}

	if m := keyRe.FindStringSubmatch(pgErr.Detail); len(m) == 2 {
		return m[1]
	}

	return ""
}
