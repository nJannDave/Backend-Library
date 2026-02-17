package service

import (
	"stmnplibrary/constanta"
	"stmnplibrary/controller/service/utils"
	"stmnplibrary/domain/entity"
	"stmnplibrary/domain/interface/repository"
	"stmnplibrary/domain/interface/service"
	"stmnplibrary/dto"
	"stmnplibrary/security"
	"strings"

	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/sync/singleflight"
)

const keyReTk = "stmnplibrary:accesstoken:id:%d"
const keyBook = "stmnplibrary:book:id:%v"
const keyBlcklist = "blacklist:accesstoken:%s"

type userService struct {
	userRepository    repository.UserRepository
	singleFlightGroup *singleflight.Group
}

func FnUserService(repo repository.UserRepository) service.UserService {
	return &userService{
		userRepository:    repo,
		singleFlightGroup: &singleflight.Group{},
	}
}

func (us *userService) getBook(ctx context.Context, key string, offset int, stop int) []dto.Books {
	id, _ := us.userRepository.RedisZR(ctx, key, offset, stop)
	var resltRds = make([]dto.Books, len(id))
	rsl, _ := us.userRepository.RedisWp(ctx, func(ctx context.Context) (interface{}, error) {
		for z, i := range id{
			us.userRepository.RedisHGetAll(ctx, keyBook + i, &resltRds[z])
		}
		return resltRds, nil
	})
	return rsl.([]dto.Books)
}

func (us *userService) setBook(ctx context.Context, keyIdx string, books []dto.Books) {
	us.userRepository.RedisWp(ctx, func(ctx context.Context) (interface{}, error) {
		for _, i := range books {
			us.userRepository.RedisZS(ctx, keyIdx, float64(time.Now().UnixNano()), i.ID)
			us.userRepository.RedisHSET(ctx, keyBook + strconv.Itoa(i.ID), i)
		}
		return nil, nil
	})
}

func (us *userService) RateLimiter(ctx context.Context, ip string) error {
	var key = "rate-limiter:ip:" + ip
	if err := us.userRepository.RateLimiter(ctx, key); err != nil {
		return utils.ValidateErrTw(err, "service - rate_limiter: %w")
	}
	return nil
}

func (us *userService) CheckAccTkn(ctx context.Context, acc string) error {
	var keyBlck = fmt.Sprintf(keyBlcklist, acc)
	rslt, err := us.userRepository.RedisGet(ctx, keyBlck)
	if err != nil {
		if strings.Contains(err.Error(), "no data found") {
			return nil
		}
		return fmt.Errorf("internal server error. something wrong: %w", err)
	}
	if rslt != nil {
		return fmt.Errorf("this access token has been blacklist")
	}
	return nil
}

func (us *userService) Logout(ctx context.Context) error {
	const errIntrnl = "service - logout: %w"
	var (
		msg      = "please login"
		duration = time.Now().Add(3 * time.Minute)
		ttl      = time.Until(duration)
	)
	userId, ok := ctx.Value(constanta.UI).(int)
	if !ok {
		return errors.New(msg)
	}
	token, ok := ctx.Value(constanta.TokenA).(string)
	if !ok {
		return errors.New(msg)
	}
	var (
		keyDel = fmt.Sprintf(keyReTk, userId)
		keySet = fmt.Sprintf(keyBlcklist, token)
	)
	return us.userRepository.RedisWtx(ctx, func(ctx context.Context) error {
		if err := us.userRepository.RedisDel(ctx, keyDel); err != nil {
			return utils.ValidateErrTw(err, errIntrnl)
		}
		if err := us.userRepository.RedisSet(ctx, keySet, []byte(token), ttl); err != nil {
			return utils.ValidateErrTw(err, errIntrnl)
		}
		return nil
	})
}

func (us *userService) Register(ctx context.Context, data *dto.Students) ([]string, error) {
	const errIntrnl = "service - register: %w"
	var errMsg []string
	realData := &entity.Students{
		NIS: data.NIS,
		PersonalInfo: entity.PersonalInfo{
			Name:        data.Name,
			Email:       data.Email,
			PhoneNumber: data.PhoneNumber,
		},
		Password: data.Password,
		AcademicInfo: entity.AcademicInfo{
			Class:    data.Class,
			SubClass: data.SubClass,
			Batch:    data.Batch,
			Major:    data.Major,
		},
	}
	if errNis := us.userRepository.GetNIS(ctx, realData.NIS); errNis != nil {
		if errValN := utils.ValidateErr(errNis, "registered", &errMsg); errValN != nil {
			return nil, errValN
		}
	}
	if errEma := us.userRepository.GetEmail(ctx, realData.PersonalInfo.Email); errEma != nil {
		if errValE := utils.ValidateErr(errEma, "used", &errMsg); errValE != nil {
			return nil, errValE
		}
	}
	realData.PersonalInfo.ValidatePN(&errMsg)
	realData.AcademicInfo.ValidateClass(&errMsg)
	if len(errMsg) > 0 {
		return errMsg, nil
	}
	hashPass, errPass := security.HashPassword(realData.Password)
	if errPass != nil {
		return nil, fmt.Errorf(errIntrnl, errPass)
	}
	realData.Password = hashPass
	if err := us.userRepository.Register(ctx, realData); err != nil {
		return nil, utils.ValidateErrTw(err, errIntrnl)
	}
	return nil, nil
}

func (us *userService) GetBooks(ctx context.Context, page int) ([]dto.Books, error) {
	const keyBooks = "stmnplibrary:books:all:page:%d"
	var (
		limit  = 35
		offset = ((page - 1) * limit)
		stop   = offset + limit - 1
	)

	result, err, _ := us.singleFlightGroup.Do(keyBooks, func() (interface{}, error) {
		rsl := us.getBook(ctx, fmt.Sprintf(keyBooks, page), offset, stop)
		if rsl != nil && len(rsl) > 0 {
			return rsl, nil
		}
		result, err := us.userRepository.GetBooks(ctx, offset)
		if err != nil {
			return nil, utils.ValidateErrTw(err, "service - get_books: %w")
		}
		books := utils.BooksMapper(result)
		us.setBook(ctx, keyBooks, books)
		return books, nil
	})
	if err != nil {
		return nil, err
	}
	return result.([]dto.Books), nil
}

func (us *userService) GetBooksByAuthor(ctx context.Context, author string, page int) ([]dto.Books, error) {
	const keyA = "stmnplibrary:getbooks:author:%s:page:%d"
	var (
		limit     = 35
		offset    = (page - 1) * limit
		stop      = offset + limit - 1
		keyAuthor = keyA + author + strconv.Itoa(page)
	)

	rsl := us.getBook(ctx, keyAuthor, offset, stop)
	if rsl != nil && len(rsl) > 0 {
		return rsl, nil
	}

	result, err := us.userRepository.GetBooksByAuthor(ctx, author, offset)
	if err != nil {
		return nil, utils.ValidateErrTw(err, "service - get_books_by_author: %w")
	}
	books := utils.BooksMapper(result)

	us.setBook(ctx, fmt.Sprintf(keyAuthor, author), books)

	return books, nil
}

func (us *userService) GetBooksByCategory(ctx context.Context, category []string, page int) ([]dto.Books, error) {
	const keyCategory = "stmnplibrary:category:%s:page:%d"
	var (
		limit       = 35
		offset      = (page - 1) * limit
		stop        = offset + limit - 1
		id          []string
		encountered = map[string]interface{}{}
	)

	result, err, _ := us.singleFlightGroup.Do(keyCategory, func() (interface{}, error) {
		x, _ := us.userRepository.RedisWp(ctx, func(ctx context.Context) (interface{}, error) {
			for _, c := range category {
				idb, _ := us.userRepository.RedisZR(ctx, keyCategory + c + strconv.Itoa(page), offset, stop)
				for _, i := range idb {
					if _, ok := encountered[i]; !ok {
						id = append(id, i)
						encountered[i] = struct{}{}
					}
				}
			}
			var resltRds = make([]dto.Books, len(id))
			for z, i := range id {
				us.userRepository.RedisHGetAll(ctx, keyBook + i, &resltRds[z])
			}
			return resltRds, nil
		})
		rsl, _ := x.([]dto.Books)
		if rsl != nil && len(rsl) > 0 && rsl[0].ID != 0 {
			fmt.Printf("ini dari cache: debug")
			return rsl, nil
		}

		result, err := us.userRepository.GetBooksByCategory(ctx, category, offset)
		if err != nil {
			return nil, utils.ValidateErrTw(err, "service - get_books_by_category: %w")
		}
		books := utils.BooksMapper(result)

		us.userRepository.RedisWp(ctx, func(ctx context.Context) (interface{}, error) {
			for _, i := range books {
				for _, c := range category {
					us.userRepository.RedisZS(ctx, keyCategory + c + strconv.Itoa(page), float64(time.Now().UnixNano()), i.ID)
				}
				us.userRepository.RedisHSET(ctx, keyBook + strconv.Itoa(i.ID), i)
			}
			return nil, nil
		})
		
		return books, nil
	})
	if err != nil {
		return nil, err
	}
	return result.([]dto.Books), nil
}

func (us *userService) Loan(ctx context.Context, loanInfo dto.Loan) error {
	const errIntrnl = "service - loan: %w"
	idUser, ok := ctx.Value(constanta.UI).(int)
	if !ok {
		return fmt.Errorf("please login")
	}
	return us.userRepository.WithContext(ctx, func(ctx context.Context) error {
		if err := us.userRepository.CheckLoan(ctx, loanInfo.ID, idUser); err != nil {
			return utils.ValidateErrLoan(err, "")
		}
		entityLoanData := &entity.Loan{
			IdUser: idUser,
			IdBook: loanInfo.ID,
		}
		if err := entityLoanData.ValidateDateFormat(loanInfo.ReturnedAt); err != nil {
			return err
		}
		if err := entityLoanData.ValidateDate(); err != nil {
			return err
		}
		if err := us.userRepository.CreateLoan(ctx, *entityLoanData); err != nil {
			return utils.ValidateErrLoan(err, "")
		}
		if err := us.userRepository.UpdateBookStock(ctx, loanInfo.ID); err != nil {
			return utils.ValidateErrLoan(err, "book")
		}
		if err := us.userRepository.UpdateLimitLoan(ctx, idUser); err != nil {
			return utils.ValidateErrLoan(err, "user")
		}
		return nil
	})
}
