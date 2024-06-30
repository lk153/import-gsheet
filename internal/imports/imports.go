package imports

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lk153/gsheet-go/lib"

	config2 "github.com/lk153/import-gsheet/internal/config"
	"github.com/lk153/import-gsheet/internal/models"
	"github.com/lk153/import-gsheet/lib/db"
	"github.com/lk153/import-gsheet/utils"
)

func Import() {
	database := db.Open(config2.GetCfg())
	sqlxDB := sqlx.NewDb(database, "mysql")
	dbInstance := sqlxDB.Unsafe()

	/*Get Categories map for later updates*/
	// cateMap := getMostChildCateMap(dbInstance)
	// fmt.Println("cateMap", cateMap)
	// os.Exit(1)

	srv, err := lib.NewGsheetServiceV2()
	if err != nil {
		fmt.Println("Cannot connect Gsheet!")
		return
	}

	// spreadsheetID := "1JhddKKmez9thv8lKmybM09o_MiEbAnPlxitIAqeyK04"
	spreadsheetID := "1c-onQeYHmvc-EPkrJDU-WyAydbCAA1ng6hXCgdYiqqg"
	readRange := "'To Update on DB'!A3:AR3"
	values := srv.ReadSheet(spreadsheetID, readRange)
	for idx, row := range values {
		fmt.Println(utils.Info("=========================================================== " + strconv.Itoa(idx) + " ==========================================================="))
		fmt.Println("Data: ", strings.Join(row, " | "))
		fmt.Println()
		fmt.Println("*----------------------------------------------------------DB----------------------------------------------------------*")
		if err = BulkUpdate(dbInstance, row); err != nil {
			fmt.Println("ERROR: ", err.Error())
		}
		fmt.Println("*----------------------------------------------------------------------------------------------------------------------*")
		fmt.Println()
		fmt.Println()
	}
}

func BulkUpdate(dbInstance *sqlx.DB, row []string) (err error) {
	supplierID, err := strconv.ParseInt(row[0], 10, 64)
	if err != nil {
		fmt.Println("BulkUpdate:ERROR: ", err.Error(), row[0])
		return errors.New(fmt.Sprintf("BulkUpdate:ERROR: %v - %s", err.Error(), row[0]))
	}

	if supplierID == 0 {
		fmt.Println("BulkUpdate:ERROR: supplierID is empty")
		return errors.New(fmt.Sprintf("BulkUpdate:ERROR: supplierID is empty"))
	}

	/*Init Supplier and related models*/
	supplierBean := &models.Supplier{Id: supplierID}
	supplierDetailBean := &models.SupplierDetail{SupplierId: supplierID}
	bankAccountBean := &models.BankAccountDetails{SupplierId: supplierID}

	/*Prepare Supplier updation query*/
	updateSupplierQuery := prepareSupplierUpdateSQL(supplierBean, row)
	updateSupplierDetailQuery := prepareSupplierDetailUpdateSQL(supplierDetailBean, row)

	/*Execute Supplier updation query on DB*/
	tx := dbInstance.MustBegin()

	if err = execSupplierUpdate(tx, updateSupplierQuery, supplierBean); err != nil {
		if err = tx.Rollback(); err != nil {
			fmt.Println(utils.Fatal("execSupplierUpdate: Rollback Failed: ", err.Error()))
		}
	}

	if err = execSupplierDetailUpdate(tx, updateSupplierDetailQuery, supplierDetailBean); err != nil {
		if err = tx.Rollback(); err != nil {
			fmt.Println(utils.Fatal("execSupplierDetailUpdate: Rollback Failed: ", err.Error()))
		}
	}

	if isBankInformationExist(dbInstance, supplierID) {
		updateBankAccountQuery := prepareBankAccountDetailUpdateSQL(bankAccountBean, row)
		if err = execBankAccountUpdate(tx, updateBankAccountQuery, bankAccountBean); err != nil {
			if err = tx.Rollback(); err != nil {
				fmt.Println(utils.Fatal("execBankAccountUpdate: Rollback Failed: ", err.Error()))
			}
		}
	} else {
		insertBankAccountQuery := prepareBankAccountDetailInsertSQL(bankAccountBean, row)
		if err = execBankAccountInsert(tx, insertBankAccountQuery, bankAccountBean); err != nil {
			if err = tx.Rollback(); err != nil {
				fmt.Println(utils.Fatal("execBankAccountInsert: Rollback Failed: ", err.Error()))
			}
		}
	}

	if err = tx.Commit(); err != nil {
		fmt.Println(utils.Fatal("Cannot commit DB transaction: ", err.Error()))
	}

	// if !checkSupplierCategoryUpdated(dbInstance, supplierID, row[28]) {
	// 	fmt.Println(utils.Fatal("checkSupplierCategoryUpdated: NOT Existed:", supplierID))
	// }

	return
}

func execSupplierUpdate(tx *sqlx.Tx, updateSupplierQuery string, supplierBean *models.Supplier) (err error) {
	var result sql.Result
	if result, err = tx.NamedExec(updateSupplierQuery, supplierBean); err != nil {
		fmt.Println(utils.Fatal("execSupplierUpdate: Error: ", err.Error()))
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		fmt.Println(utils.Fatal("execSupplierUpdate: RowsAffected: ", err.Error()))
		return
	}

	fmt.Println(utils.Debug("execSupplierUpdate: Affected: ", affected))
	return
}

func execSupplierDetailUpdate(tx *sqlx.Tx, updateSupplierDetailQuery string, supplierDetailBean *models.SupplierDetail) (err error) {
	var result sql.Result
	if result, err = tx.NamedExec(updateSupplierDetailQuery, supplierDetailBean); err != nil {
		fmt.Println(utils.Fatal("execSupplierDetailUpdate: Error: ", err.Error()))
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		fmt.Println(utils.Fatal("execSupplierDetailUpdate: RowsAffected: ", err.Error()))
		return
	}

	fmt.Println(utils.Debug("execSupplierDetailUpdate: Affected: ", affected))
	return
}

func execBankAccountUpdate(tx *sqlx.Tx, updateBankAccountQuery string, bankAccountBean *models.BankAccountDetails) (err error) {
	var result sql.Result
	if result, err = tx.NamedExec(updateBankAccountQuery, bankAccountBean); err != nil {
		fmt.Println(utils.Fatal("execBankAccountUpdate: Error: ", err.Error()))
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		fmt.Println(utils.Fatal("execBankAccountUpdate: RowsAffected: ", err.Error()))
		return
	}

	fmt.Println(utils.Debug("execBankAccountUpdate: Affected: ", affected))
	return
}

func execBankAccountInsert(tx *sqlx.Tx, insertBankAccountQuery string, bankAccountBean *models.BankAccountDetails) (err error) {
	var result sql.Result
	if result, err = tx.NamedExec(insertBankAccountQuery, bankAccountBean); err != nil {
		fmt.Println(utils.Fatal("execBankAccountInsert: Error: ", err.Error()))
		return
	}

	lastInsertId, err := result.LastInsertId()
	if err != nil {
		fmt.Println(utils.Fatal("execBankAccountInsert: LastInsertId: ", err.Error()))
		return
	}

	fmt.Println(utils.Debug("execBankAccountInsert: Inserted ID: ", lastInsertId))
	return
}

func prepareSupplierUpdateSQL(s *models.Supplier, row []string) (updateSupplierQuery string) {
	updateSupplierQuery = `UPDATE suppliers SET %s WHERE id = :id`
	setFields := []string{}

	if len(strings.TrimSpace(row[1])) != 0 {
		s.Entity = strings.TrimSpace(row[1])
		setFields = append(setFields, `entity = :entity`)
	}
	if len(strings.TrimSpace(row[2])) != 0 {
		s.CompanyName = strings.TrimSpace(row[2])
		setFields = append(setFields, `company_name = :company_name`)
	}
	if len(strings.TrimSpace(row[3])) != 0 {
		s.AlternateCompanyName = sql.NullString{String: strings.TrimSpace(row[3]), Valid: true}
		setFields = append(setFields, `alternate_company_name = :alternate_company_name`)
	}
	if len(strings.TrimSpace(row[8])) != 0 {
		s.City = sql.NullString{String: strings.TrimSpace(row[8]), Valid: true}
		setFields = append(setFields, `city = :city`)
	}
	if len(strings.TrimSpace(row[9])) != 0 {
		s.LocationRegion = sql.NullString{String: strings.TrimSpace(row[9]), Valid: true}
		setFields = append(setFields, `location_region = :location_region`)
	}
	if len(strings.TrimSpace(row[10])) != 0 {
		s.LegalPerson = sql.NullString{String: strings.TrimSpace(row[10]), Valid: true}
		setFields = append(setFields, `legal_person = :legal_person`)
	}
	if len(strings.TrimSpace(row[11])) != 0 {
		s.LegalPersonId = sql.NullString{String: strings.TrimSpace(row[11]), Valid: true}
		setFields = append(setFields, `legal_person_id = :legal_person_id`)
	}
	if len(strings.TrimSpace(row[13])) != 0 {
		s.NumberOfEmployeesRangeID = sql.NullInt64{Int64: getNumberEmployeeRangeID(row[13]), Valid: true}
		setFields = append(setFields, `number_of_employees_range_id = :number_of_employees_range_id`)
	}
	if len(strings.TrimSpace(row[14])) != 0 {
		s.PassedVetting = sql.NullString{String: strings.TrimSpace(row[14]), Valid: true}
		setFields = append(setFields, `passed_vetting = :passed_vetting`)
	}
	if len(strings.TrimSpace(row[15])) != 0 {
		s.VettingInfoUrl = sql.NullString{String: strings.TrimSpace(row[15]), Valid: true}
		setFields = append(setFields, `vetting_info_url = :vetting_info_url`)
	}
	if len(strings.TrimSpace(row[16])) != 0 {
		s.ContactPerson = strings.TrimSpace(row[16])
		setFields = append(setFields, `contact_person = :contact_person`)
	}
	if len(strings.TrimSpace(row[17])) != 0 {
		s.ContactNumber = strings.TrimSpace(row[17])
		setFields = append(setFields, `contact_number = :contact_number`)
	}
	if len(strings.TrimSpace(row[18])) != 0 {
		s.SocialNetworkId = sql.NullString{String: strings.TrimSpace(row[18]), Valid: true}
		setFields = append(setFields, `social_network_id = :social_network_id`)
	}

	updateSupplierQuery = fmt.Sprintf(updateSupplierQuery, strings.Join(setFields, ", "))
	return
}

func prepareSupplierDetailUpdateSQL(sd *models.SupplierDetail, row []string) (updateSupplierDetailQuery string) {
	updateSupplierDetailQuery = `UPDATE supplier_details SET %s WHERE supplier_id = :supplier_id`
	setFields := []string{}
	if len(strings.TrimSpace(row[4])) != 0 {
		sd.BusinessRegistrationNumber = sql.NullString{String: strings.TrimSpace(row[4]), Valid: true}
		setFields = append(setFields, `business_registration_number = :business_registration_number`)
	}
	if len(strings.TrimSpace(row[5])) != 0 {
		sd.RegisteredBusinessAddress = sql.NullString{String: strings.TrimSpace(row[5]), Valid: true}
		setFields = append(setFields, `registered_business_address = :registered_business_address`)
	}
	if len(strings.TrimSpace(row[6])) != 0 {
		sd.SupplierAddress = sql.NullString{String: strings.TrimSpace(row[6]), Valid: true}
		setFields = append(setFields, `supplier_address = :supplier_address`)
	}
	if len(strings.TrimSpace(row[7])) != 0 {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(row[7]))
		if err == nil {
			sd.DateOfEstablishment = sql.NullTime{Time: t, Valid: true}
			setFields = append(setFields, `date_of_establishment = :date_of_establishment`)
		}
	}
	if len(strings.TrimSpace(row[12])) != 0 {
		i, err := strconv.ParseInt(row[12], 10, 64)
		if err == nil {
			sd.PaidUpCapitalRMB = sql.NullInt64{Int64: i, Valid: true}
			setFields = append(setFields, `paid_up_capital_in_rmb = :paid_up_capital_in_rmb`)
		}
	}
	if len(strings.TrimSpace(row[19])) != 0 {
		sd.EmailAddress = sql.NullString{String: strings.TrimSpace(row[19]), Valid: true}
		setFields = append(setFields, `email_address = :email_address`)
	}
	if len(strings.TrimSpace(row[20])) != 0 {
		sd.SupplierWebsiteURL = sql.NullString{String: strings.TrimSpace(row[20]), Valid: true}
		setFields = append(setFields, `supplier_website_url = :supplier_website_url`)
	}
	if len(strings.TrimSpace(row[21])) != 0 {
		sd.SupplierType = sql.NullString{String: strings.TrimSpace(row[21]), Valid: true}
		setFields = append(setFields, `supplier_type = :supplier_type`)
	}
	if len(strings.TrimSpace(row[22])) != 0 {
		i, err := strconv.ParseInt(row[22], 10, 16)
		if err == nil {
			sd.BrandedGoods = int16(i)
			setFields = append(setFields, `branded_goods = :branded_goods`)
		}
	}
	if len(strings.TrimSpace(row[23])) != 0 {
		sd.BrandCheckID = sql.NullString{String: strings.TrimSpace(row[23]), Valid: true}
		setFields = append(setFields, `brand_check_id = :brand_check_id`)
	}
	if len(strings.TrimSpace(row[25])) != 0 {
		sd.OriginSource = strings.TrimSpace(row[25])
		setFields = append(setFields, `origin_source = :origin_source`)
	}
	if len(strings.TrimSpace(row[26])) != 0 {
		switch strings.ToUpper(row[26]) {
		case "YES":
			sd.HonestCivilDebtor = sql.NullBool{Bool: true, Valid: true}
		case "NO":
			sd.HonestCivilDebtor = sql.NullBool{Bool: false, Valid: true}
		}
		setFields = append(setFields, `honest_civil_debtor = :honest_civil_debtor`)
	}
	if len(strings.TrimSpace(row[27])) != 0 {
		switch strings.ToUpper(row[27]) {
		case "YES":
			sd.InvoiceUnderNinja = sql.NullBool{Bool: true, Valid: true}
		case "NO":
			sd.InvoiceUnderNinja = sql.NullBool{Bool: false, Valid: true}
		}
		setFields = append(setFields, `invoice_under_ninja = :invoice_under_ninja`)
	}

	updateSupplierDetailQuery = fmt.Sprintf(updateSupplierDetailQuery, strings.Join(setFields, ", "))
	return
}

func getNumberEmployeeRangeID(name string) int64 {
	switch name {
	case "<50":
		return 1
	case "50-99":
		return 2
	case "100-499":
		return 3
	case "500-999":
		return 4
	case "1000-4999":
		return 5
	case ">=5000":
		return 6
	default:
		return 0
	}
}

func prepareBankAccountDetailInsertSQL(ba *models.BankAccountDetails, row []string) (insertBankAccountQuery string) {
	insertBankAccountQuery = `INSERT INTO bank_account_details (%s) VALUES (%s)`
	setFields := []string{`:supplier_id`}
	columns := []string{`supplier_id`}
	if len(strings.TrimSpace(row[29])) != 0 {
		ba.AccountType = sql.NullString{String: strings.TrimSpace(row[29]), Valid: true}
		setFields = append(setFields, `:account_type`)
		columns = append(columns, `account_type`)
	}
	if len(strings.TrimSpace(row[30])) != 0 {
		ba.AccountHolderName = sql.NullString{String: strings.TrimSpace(row[30]), Valid: true}
		setFields = append(setFields, `:account_holder_name`)
		columns = append(columns, `account_holder_name`)
	}
	if len(strings.TrimSpace(row[31])) != 0 {
		ba.AccountNumber = sql.NullString{String: strings.TrimSpace(row[31]), Valid: true}
		setFields = append(setFields, `:account_number`)
		columns = append(columns, `account_number`)
	}
	if len(strings.TrimSpace(row[32])) != 0 {
		ba.BankName = sql.NullString{String: strings.TrimSpace(row[32]), Valid: true}
		setFields = append(setFields, `:bank_name`)
		columns = append(columns, `bank_name`)
	}
	if len(strings.TrimSpace(row[33])) != 0 {
		ba.SwiftCode = sql.NullString{String: strings.TrimSpace(row[33]), Valid: true}
		setFields = append(setFields, `:swift_code`)
		columns = append(columns, `swift_code`)
	}
	if len(strings.TrimSpace(row[34])) != 0 {
		ba.BankAddress = sql.NullString{String: strings.TrimSpace(row[34]), Valid: true}
		setFields = append(setFields, `:bank_address`)
		columns = append(columns, `bank_address`)
	}
	if len(strings.TrimSpace(row[35])) != 0 {
		ba.SupplierCompanyAddress = sql.NullString{String: strings.TrimSpace(row[35]), Valid: true}
		setFields = append(setFields, `:supplier_company_address`)
		columns = append(columns, `supplier_company_address`)
	}

	insertBankAccountQuery = fmt.Sprintf(insertBankAccountQuery, strings.Join(columns, ", "), strings.Join(setFields, ", "))
	return
}

func prepareBankAccountDetailUpdateSQL(ba *models.BankAccountDetails, row []string) (updateBankAccountQuery string) {
	updateBankAccountQuery = `UPDATE bank_account_details SET %s WHERE supplier_id = :supplier_id`
	setFields := []string{}
	if len(strings.TrimSpace(row[29])) != 0 {
		ba.AccountType = sql.NullString{String: strings.TrimSpace(row[29]), Valid: true}
		setFields = append(setFields, `account_type = :account_type`)
	}
	if len(strings.TrimSpace(row[30])) != 0 {
		ba.AccountHolderName = sql.NullString{String: strings.TrimSpace(row[30]), Valid: true}
		setFields = append(setFields, `account_holder_name = :account_holder_name`)
	}
	if len(strings.TrimSpace(row[31])) != 0 {
		ba.AccountNumber = sql.NullString{String: strings.TrimSpace(row[31]), Valid: true}
		setFields = append(setFields, `account_number = :account_number`)
	}
	if len(strings.TrimSpace(row[32])) != 0 {
		ba.BankName = sql.NullString{String: strings.TrimSpace(row[32]), Valid: true}
		setFields = append(setFields, `bank_name = :bank_name`)
	}
	if len(strings.TrimSpace(row[33])) != 0 {
		ba.SwiftCode = sql.NullString{String: strings.TrimSpace(row[33]), Valid: true}
		setFields = append(setFields, `swift_code = :swift_code`)
	}
	if len(strings.TrimSpace(row[34])) != 0 {
		ba.BankAddress = sql.NullString{String: strings.TrimSpace(row[34]), Valid: true}
		setFields = append(setFields, `bank_address = :bank_address`)
	}
	if len(strings.TrimSpace(row[35])) != 0 {
		ba.SupplierCompanyAddress = sql.NullString{String: strings.TrimSpace(row[35]), Valid: true}
		setFields = append(setFields, `supplier_company_address = :supplier_company_address`)
	}

	updateBankAccountQuery = fmt.Sprintf(updateBankAccountQuery, strings.Join(setFields, ", "))
	return
}

func getMostChildCateMap(dbInstance *sqlx.DB) map[string]uint {
	type Cate struct {
		Id   uint   `db:"category_id"`
		Name string `db:"name"`
	}

	cate := Cate{}
	cateMap := map[string]uint{}
	rows, err := dbInstance.Queryx(`SELECT c.* FROM categories c 
		WHERE c.category_id NOT IN (
			SELECT c2.parent_id FROM categories c2 
			WHERE c2.deleted_at IS NULL
			GROUP BY c2.parent_id
		) AND c.deleted_at IS NULL;`)
	if err != nil {
		log.Panicln(err)
	}

	for rows.Next() {
		err := rows.StructScan(&cate)
		if err != nil {
			log.Panicln(err)
		}

		cateMap[cate.Name] = cate.Id
	}

	return cateMap
}

func isBankInformationExist(dbInstance *sqlx.DB, supplierID int64) bool {
	sql := `SELECT bad.*
	FROM bank_account_details bad
	WHERE bad.supplier_id = ? AND bad.deleted_at IS NULL;`
	rows, err := dbInstance.Queryx(sql, supplierID)
	if err != nil {
		log.Panicln(err)
		return false
	}

	return rows.Next()
}

func checkSupplierCategoryUpdated(dbInstance *sqlx.DB, supplierID int64, reqCateName string) bool {
	if len(strings.TrimSpace(reqCateName)) == 0 {
		return true
	}

	slc := strings.Split(reqCateName, ",")
	for i := range slc {
		slc[i] = strings.TrimSpace(slc[i])
	}

	slc = append(slc, reqCateName)

	type SupplierCate struct {
		Id   uint   `db:"category_id"`
		Name string `db:"name"`
	}

	query, args, err := sqlx.In(`SELECT sc.category_id, c.name FROM supplier_categories sc 
	JOIN categories c ON c.category_id = sc.category_id AND c.deleted_at IS NULL AND c.name IN (?)
	WHERE sc.supplier_id = ?
	AND sc.deleted_at IS NULL;`, slc, supplierID)
	if err != nil {
		log.Panicln(err)
	}

	query = dbInstance.Rebind(query)
	supCate := SupplierCate{}
	rows, err := dbInstance.Queryx(query, args...)
	if err != nil {
		log.Panicln(err)
	}

	isExisted := false
	for rows.Next() {
		err := rows.StructScan(&supCate)
		if err != nil {
			log.Panicln(err)
		}

		isExisted = true
	}

	return isExisted
}
