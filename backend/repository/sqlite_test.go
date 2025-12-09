package repository

import (
	"os"
	"testing"

	"wechat-notification/models"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Helper function to create a test repository with a temporary database
func setupTestRepo(t *testing.T) (*SQLiteRepository, func()) {
	tmpFile, err := os.CreateTemp("", "test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	repo, err := NewSQLiteRepository(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to create repository: %v", err)
	}

	cleanup := func() {
		repo.Close()
		os.Remove(tmpFile.Name())
	}

	return repo, cleanup
}

// Generator for valid OpenID strings (non-empty alphanumeric)
func genValidOpenID() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) <= 64
	})
}

// Generator for valid name strings (non-empty)
func genValidName() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) <= 100
	})
}

// **Feature: wechat-notification, Property 7: 接收者添加持久化**
// *对于任意* 有效的接收者数据（有效 OpenID 和名称），添加后查询应能获取到该接收者
// **验证: 需求 3.2, 5.1**
func TestProperty7_RecipientAddPersistence(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("Adding a recipient should make it retrievable", prop.ForAll(
		func(openID, name string) bool {
			repo, cleanup := setupTestRepo(t)
			defer cleanup()

			recipient := &models.Recipient{
				OpenID: openID,
				Name:   name,
			}

			// Create the recipient
			err := repo.Create(recipient)
			if err != nil {
				return false
			}

			// Verify it was assigned an ID
			if recipient.ID == 0 {
				return false
			}

			// Retrieve by ID
			retrieved, err := repo.GetByID(recipient.ID)
			if err != nil {
				return false
			}

			// Verify the data matches
			return retrieved.OpenID == openID && retrieved.Name == name
		},
		genValidOpenID(),
		genValidName(),
	))

	properties.TestingRun(t)
}


// **Feature: wechat-notification, Property 8: 重复 OpenID 拒绝**
// *对于任意* 已存在的 OpenID，尝试添加具有相同 OpenID 的接收者应被拒绝
// **验证: 需求 3.3**
func TestProperty8_DuplicateOpenIDRejection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("Adding a recipient with duplicate OpenID should be rejected", prop.ForAll(
		func(openID, name1, name2 string) bool {
			repo, cleanup := setupTestRepo(t)
			defer cleanup()

			// Create first recipient
			recipient1 := &models.Recipient{
				OpenID: openID,
				Name:   name1,
			}
			err := repo.Create(recipient1)
			if err != nil {
				return false
			}

			// Try to create second recipient with same OpenID
			recipient2 := &models.Recipient{
				OpenID: openID,
				Name:   name2,
			}
			err = repo.Create(recipient2)

			// Should return ErrDuplicateOpenID
			return err == ErrDuplicateOpenID
		},
		genValidOpenID(),
		genValidName(),
		genValidName(),
	))

	properties.TestingRun(t)
}

// **Feature: wechat-notification, Property 9: 接收者删除**
// *对于任意* 存在的接收者，删除后查询应无法获取到该接收者
// **验证: 需求 3.4**
func TestProperty9_RecipientDeletion(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("Deleting a recipient should make it unretrievable", prop.ForAll(
		func(openID, name string) bool {
			repo, cleanup := setupTestRepo(t)
			defer cleanup()

			// Create a recipient
			recipient := &models.Recipient{
				OpenID: openID,
				Name:   name,
			}
			err := repo.Create(recipient)
			if err != nil {
				return false
			}

			// Delete the recipient
			err = repo.Delete(recipient.ID)
			if err != nil {
				return false
			}

			// Try to retrieve - should return ErrNotFound
			_, err = repo.GetByID(recipient.ID)
			return err == ErrNotFound
		},
		genValidOpenID(),
		genValidName(),
	))

	properties.TestingRun(t)
}

// **Feature: wechat-notification, Property 10: 接收者更新**
// *对于任意* 存在的接收者和新名称，更新后查询应返回新名称
// **验证: 需求 3.5**
func TestProperty10_RecipientUpdate(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("Updating a recipient name should persist the new name", prop.ForAll(
		func(openID, originalName, newName string) bool {
			repo, cleanup := setupTestRepo(t)
			defer cleanup()

			// Create a recipient
			recipient := &models.Recipient{
				OpenID: openID,
				Name:   originalName,
			}
			err := repo.Create(recipient)
			if err != nil {
				return false
			}

			// Update the name
			recipient.Name = newName
			err = repo.Update(recipient)
			if err != nil {
				return false
			}

			// Retrieve and verify
			retrieved, err := repo.GetByID(recipient.ID)
			if err != nil {
				return false
			}

			return retrieved.Name == newName
		},
		genValidOpenID(),
		genValidName(),
		genValidName(),
	))

	properties.TestingRun(t)
}

// **Feature: wechat-notification, Property 12: 数据持久化往返**
// *对于任意* 接收者数据，保存到数据库后重新加载应得到等价的数据
// **验证: 需求 5.2**
func TestProperty12_DataPersistenceRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("Saving and loading recipient data should preserve all fields", prop.ForAll(
		func(openID, name string) bool {
			repo, cleanup := setupTestRepo(t)
			defer cleanup()

			// Create a recipient
			original := &models.Recipient{
				OpenID: openID,
				Name:   name,
			}
			err := repo.Create(original)
			if err != nil {
				return false
			}

			// Retrieve via GetAll
			all, err := repo.GetAll()
			if err != nil {
				return false
			}

			if len(all) != 1 {
				return false
			}

			retrieved := all[0]

			// Verify all fields match
			return retrieved.ID == original.ID &&
				retrieved.OpenID == original.OpenID &&
				retrieved.Name == original.Name &&
				!retrieved.CreatedAt.IsZero() &&
				!retrieved.UpdatedAt.IsZero()
		},
		genValidOpenID(),
		genValidName(),
	))

	properties.TestingRun(t)
}
