package bookstore

import "context"

type BookstoreService struct {
	storeConn StoreConn
}

// BookstoreService uses StoreConn interface because it needs transactions
func NewBookstoreService(storeConn StoreConn) *BookstoreService {
	return &BookstoreService{
		storeConn: storeConn,
	}
}

func (s *BookstoreService) LollifyFirstAuthor(ctx context.Context) {
	tx, _ := s.storeConn.BeginTx(nil)

	name := loadAuthorNameHelper(ctx, s.storeConn)
	tx.UpdateAuthor(ctx, &UpdateAuthorReq{ID: 1, Name: "lol " + name})

	tx.Commit()
}

// This helper method uses Store interface since it does not care about transactions
func loadAuthorNameHelper(ctx context.Context, store Store) string {
	authors, _ := store.GetAuthors(ctx, GetAuthorsReq{ID: 1})
	return authors[0].Name
}
