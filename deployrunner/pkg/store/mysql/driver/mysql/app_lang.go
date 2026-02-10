package store

import (
	"context"

	"taskrunner/pkg/sql-driver/driver"
	"taskrunner/trait"
)

func (c *SQLCursor) InsertAppLang(ctx context.Context, lang, aname, alias, zone string) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.InsertAppLang, lang, zone+aname, alias)
	return err
}

func (c *SQLCursor) UpdateAPPLang(ctx context.Context, lang, aname, alias, zone string) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.UpdateAppLang, alias, lang, zone+aname)
	return err
}

func (c *SQLCursor) GetAppLang(ctx context.Context, lang, aname, zone string) (string, *trait.Error) {
	row := c.QueryRowContext(ctx, c.stmt.GetAppLang, lang, zone+aname)
	alias := aname
	err := row.Scan(&alias)
	return alias, err
}

func (tx *TX) InsertAppLang(ctx context.Context, lang, aname, alias, zone string) *trait.Error {
	cur, err := tx.GetAppLang(ctx, lang, aname, zone)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		return tx.SQLCursor.InsertAppLang(ctx, lang, aname, alias, zone)
	}
	if cur != alias {
		return tx.UpdateAPPLang(ctx, lang, aname, alias, zone)
	}
	return nil
}

func (s *Store) InsertAppLang(ctx context.Context, lang, aname, alias, zone string) *trait.Error {
	return driver.StoreTransactionMarco(s, func(tx driver.Transaction) *trait.Error {
		return s.beginWithTx(tx).InsertAppLang(ctx, lang, aname, alias, zone)
	})
}

func (c *SQLCursor) GetAname(lang, alias, zone string) string {
	return ""
}
