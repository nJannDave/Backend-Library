package service

import (
	"context"
	"errors"
	"fmt"
	"stmnplibrary/domain/entity"
	"stmnplibrary/domain/interface/repository"
	"stmnplibrary/domain/interface/service"
	"stmnplibrary/dto"
	"stmnplibrary/constanta"
	"stmnplibrary/controller/service/utils"
	"time"
)

type adminService struct {
	adminRepository repository.AdminRepository
}

func FnAdminService(repository repository.AdminRepository) service.AdminService {
	return &adminService{adminRepository: repository}
}

func (as *adminService) getLimitOffset(page int) (int, int) {
	var limit = 35
	return limit, ((page - 1) * limit)
}

func (as *adminService) GetLoanData(ctx context.Context, page int) ([]dto.LoanData, error) {
	var (
		keyLoanData = "stmnplibary:loandata:page:%d"
		_, offset = as.getLimitOffset(page)
	)
	result, _ := as.adminRepository.RedisGet(ctx, keyLoanData)
	rsl, _ := result.([]dto.LoanData)
	if rsl != nil && len(rsl) > 0{
		return rsl, nil
	}
	data, err := as.adminRepository.GetLoanData(ctx, offset)
	if err != nil {
		return nil, utils.ValidateErrTw(err, "service - get_loan_data: %w")
	}
	if data == nil || len(data) == 0 {
		return nil, fmt.Errorf("data doesn't exists")
	}
	loanData := utils.LoanDataMapper(data)
	as.adminRepository.RedisSet(ctx, fmt.Sprintf(keyLoanData, page), loanData, 3*time.Minute)
	return loanData, nil
}

func (as *adminService) GetLDDone(ctx context.Context, page int) ([]dto.LoanData, error) {
	var (
		keyLoanData = "stmnplibary:loandata:done:page:%d"
		_, offset = as.getLimitOffset(page)
	)
	result, _ := as.adminRepository.RedisGet(ctx, keyLoanData)
	rsl, _ := result.([]dto.LoanData)
	if rsl != nil && len(rsl) > 0{
		return rsl, nil
	}
	data, err := as.adminRepository.GetLDDone(ctx, offset)
	if err != nil {
		return nil, utils.ValidateErrTw(err, "service - get_loan_data_done: %w")
	}
	if data == nil || len(data) == 0 {
		return nil, fmt.Errorf("data doesn't exists")
	}
	loanData := utils.LoanDataMapper(data)
	as.adminRepository.RedisSet(ctx, fmt.Sprintf(keyLoanData, page), loanData, 3*time.Minute)
	return loanData, nil
}

func (as *adminService) GetLDDont(ctx context.Context, page int) ([]dto.LoanData, error) {
	var (
		keyLoanData = "stmnplibary:loandata:dont:page:%d"
		_, offset = as.getLimitOffset(page)
	)
	result, _ := as.adminRepository.RedisGet(ctx, keyLoanData)
	rsl, _ := result.([]dto.LoanData)
	if rsl != nil && len(rsl) > 0{
		return rsl, nil
	}
	data, err := as.adminRepository.GetLDDont(ctx, offset)
	if err != nil {
		return nil, utils.ValidateErrTw(err, "service - get_loan_data_dont: %w")
	}
	if data == nil || len(data) == 0 {
		return nil, fmt.Errorf("data doesn't exists")
	}
	loanData := utils.LoanDataMapper(data)
	as.adminRepository.RedisSet(ctx, fmt.Sprintf(keyLoanData, page), loanData, 3*time.Minute)
	return loanData, nil
}

func (as *adminService) AddCategory(ctx context.Context, data dto.Category) error {
	var (
		ik, ok = ctx.Value(string(constanta.IK)).(string)
		entityData = entity.Category{
			Name: data.Name,
		}
		keyIk = "idempotency:key:"+ik
	)
	if !ok {
		return errors.New("missing idempotency key")		
	}
	isNew, _ := as.adminRepository.RedisSETNX(ctx, keyIk, ik, 25*time.Minute)
	if !isNew {
		return errors.New("duplicate request")
	}
	if err := as.adminRepository.AddCategory(ctx, entityData); err != nil {
		if err := as.adminRepository.RedisDel(ctx, keyIk); err != nil {
			return errors.New("failed delete key")
		}
		return utils.ValidateErrTw(err, "service - add_category: %w")
	}
	return nil
}

func (as *adminService) AddBook(ctx context.Context, data dto.BookData) error {
	var (
		ik, ok = ctx.Value(string(constanta.IK)).(string)
		entityData = entity.BookData{
			ISBN:           data.ISBN,
			Name:           data.Name,
			Author:         data.Author,
			Publisher:      data.Publisher,
			Description:    data.Description,
			Stock:          data.Stock,
			AvailableStock: data.AvailableStock,
		}
		keyIk = "idempotency:key:"+ik
	)
	if !ok {
		return errors.New("missing idempotency key")
	}
	isNew, _ := as.adminRepository.RedisSETNX(ctx, keyIk, ik, 24*time.Hour)
	if !isNew {
		return errors.New("duplicate request")
	}
	if err := as.adminRepository.WithTx(ctx, func(ctx context.Context) error {
		var (
			errMsg        = "service - add_book: %w"
			entityConnect = make([]entity.Connections, 0, 10)
		)
		if err := as.adminRepository.AddBook(ctx, &entityData); err != nil {
			return utils.ValidateErrTw(err, errMsg)
		}
		for _, i := range data.IDCategory {
			entityConnect = append(entityConnect, entity.Connections{
				BookID:     entityData.BookID,
				IdCategory: i,
			})
		}
		if err := as.adminRepository.AddConnections(ctx, entityConnect); err != nil {
			return utils.ValidateErrTw(err, errMsg)
		}
		return nil
	}); err != nil {
		if err := as.adminRepository.RedisDel(ctx, keyIk); err != nil {
			return errors.New("failed delete key")
		}
		return err
	} else {
		return nil
	}
}

func (as *adminService) Confirm(ctx context.Context, data dto.Confirm) error {
	var (
		entityCf = entity.Confirm {
			Student: data.Student,
			ISBN: data.ISBN,
		}
		errMsg = "service - confirm loan: %w"
	)
	idStudent, err := as.adminRepository.GetStudentId(ctx, entityCf.Student)
	if err != nil {
		return utils.ValidateErrTw(err, errMsg)
	}
	idBook, err := as.adminRepository.GetBookId(ctx, entityCf.ISBN)
	if err != nil {
		return utils.ValidateErrTw(err, errMsg)
	}
	slData, err := as.adminRepository.GetStudentLoan(ctx, idStudent, idBook)
	if err != nil {
		return utils.ValidateErrTw(err, errMsg)
	}
	lds := utils.InitLD(&slData)
	lds.GiveSanctions()
	return as.adminRepository.WithTx(ctx, func(context.Context) error {
		if err := as.adminRepository.UpdateTabLoan(ctx, idStudent, idBook, *lds.Sanctions, *lds.ReturnedAt); err != nil {
			return utils.ValidateErrTw(err, errMsg)
		}
		if err := as.adminRepository.UpdateStock(ctx, idBook); err != nil {
			return utils.ValidateErrTw(err, errMsg)
		}
		if err := as.adminRepository.UpdateMaxBook(ctx, idStudent); err != nil {
			return utils.ValidateErrTw(err, errMsg)
		}
		return nil
	})
}
