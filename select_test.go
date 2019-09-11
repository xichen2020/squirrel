package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectBuilderToSql(t *testing.T) {
	subQ := Select("aa", "bb").From("dd")
	b := Select("a", "b").
		Prefix("WITH prefix AS ?", 0).
		Distinct().
		Columns("c").
		Column("IF(d IN ("+Placeholders(3)+"), 1, 0) as stat_column", 1, 2, 3).
		Column(Expr("a > ?", 100)).
		Column(Alias(Eq{"b": []int{101, 102, 103}}, "b_alias")).
		Column(Alias(subQ, "subq")).
		From("e").
		JoinClause("CROSS JOIN j1").
		Join("j2").
		LeftJoin("j3").
		RightJoin("j4").
		PreWhere("f = ?", 4).
		PreWhere(Eq{"g": 5}).
		PreWhere(map[string]interface{}{"h": 6}).
		PreWhere(Eq{"i": []int{7, 8, 9}}).
		PreWhere(Or{Expr("j = ?", 10), And{Eq{"k": 11}, Expr("true")}}).
		Where("f = ?", 4).
		Where(Eq{"g": 5}).
		Where(map[string]interface{}{"h": 6}).
		Where(Eq{"i": []int{7, 8, 9}}).
		Where(Or{Expr("j = ?", 10), And{Eq{"k": 11}, Expr("true")}}).
		GroupBy("l").
		Having("m = n").
		OrderByClause("? DESC", 1).
		OrderBy("o ASC", "p DESC").
		Limit(12).
		Offset(13).
		Suffix("FETCH FIRST ? ROWS ONLY", 14)

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql :=
		"WITH prefix AS ? " +
			"SELECT DISTINCT a, b, c, IF(d IN (?,?,?), 1, 0) as stat_column, a > ?, " +
			"(b IN (?,?,?)) AS b_alias, " +
			"(SELECT aa, bb FROM dd) AS subq " +
			"FROM e " +
			"CROSS JOIN j1 JOIN j2 LEFT JOIN j3 RIGHT JOIN j4 " +
			"PREWHERE f = ? AND g = ? AND h = ? AND i IN (?,?,?) AND (j = ? OR (k = ? AND true)) " +
			"WHERE f = ? AND g = ? AND h = ? AND i IN (?,?,?) AND (j = ? OR (k = ? AND true)) " +
			"GROUP BY l HAVING m = n ORDER BY ? DESC, o ASC, p DESC LIMIT 12 OFFSET 13 " +
			"FETCH FIRST ? ROWS ONLY"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{0, 1, 2, 3, 100, 101, 102, 103, 4, 5, 6, 7, 8, 9, 10, 11, 4, 5, 6, 7, 8, 9, 10, 11, 1, 14}
	assert.Equal(t, expectedArgs, args)
}

func TestSelectBuilderToSqlOnlyPreWhere(t *testing.T) {
	subQ := Select("aa", "bb").From("dd")
	b := Select("a", "b").
		Prefix("WITH prefix AS ?", 0).
		Distinct().
		Columns("c").
		Column("IF(d IN ("+Placeholders(3)+"), 1, 0) as stat_column", 1, 2, 3).
		Column(Expr("a > ?", 100)).
		Column(Alias(Eq{"b": []int{101, 102, 103}}, "b_alias")).
		Column(Alias(subQ, "subq")).
		From("e").
		JoinClause("CROSS JOIN j1").
		Join("j2").
		LeftJoin("j3").
		RightJoin("j4").
		PreWhere("f = ?", 4).
		PreWhere(Eq{"g": 5}).
		PreWhere(map[string]interface{}{"h": 6}).
		PreWhere(Eq{"i": []int{7, 8, 9}}).
		PreWhere(Or{Expr("j = ?", 10), And{Eq{"k": 11}, Expr("true")}}).
		GroupBy("l").
		Having("m = n").
		OrderByClause("? DESC", 1).
		OrderBy("o ASC", "p DESC").
		Limit(12).
		Offset(13).
		Suffix("FETCH FIRST ? ROWS ONLY", 14)

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql :=
		"WITH prefix AS ? " +
			"SELECT DISTINCT a, b, c, IF(d IN (?,?,?), 1, 0) as stat_column, a > ?, " +
			"(b IN (?,?,?)) AS b_alias, " +
			"(SELECT aa, bb FROM dd) AS subq " +
			"FROM e " +
			"CROSS JOIN j1 JOIN j2 LEFT JOIN j3 RIGHT JOIN j4 " +
			"PREWHERE f = ? AND g = ? AND h = ? AND i IN (?,?,?) AND (j = ? OR (k = ? AND true)) " +
			"GROUP BY l HAVING m = n ORDER BY ? DESC, o ASC, p DESC LIMIT 12 OFFSET 13 " +
			"FETCH FIRST ? ROWS ONLY"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{0, 1, 2, 3, 100, 101, 102, 103, 4, 5, 6, 7, 8, 9, 10, 11, 1, 14}
	assert.Equal(t, expectedArgs, args)
}

func TestSelectBuilderFromSelect(t *testing.T) {
	subQ := Select("c").From("d").PreWhere(Eq{"i": 0}).Where(Eq{"i": 0})
	b := Select("a", "b").FromSelect(subQ, "subq")
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "SELECT a, b FROM (SELECT c FROM d PREWHERE i = ? WHERE i = ?) AS subq"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{0, 0}
	assert.Equal(t, expectedArgs, args)
}

func TestSelectBuilderFromSelectNestedDollarPlaceholders(t *testing.T) {
	subQ := Select("c").
		From("t").
		PreWhere(Gt{"c": 1}).
		Where(Gt{"c": 1}).
		PlaceholderFormat(Dollar)
	b := Select("c").
		FromSelect(subQ, "subq").
		PreWhere(Lt{"c": 2}).
		Where(Lt{"c": 2}).
		PlaceholderFormat(Dollar)
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "SELECT c FROM (SELECT c FROM t PREWHERE c > $1 WHERE c > $2) AS subq PREWHERE c < $3 WHERE c < $4"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{1, 1, 2, 2}
	assert.Equal(t, expectedArgs, args)
}

func TestSelectBuilderToSqlErr(t *testing.T) {
	_, _, err := Select().From("x").ToSql()
	assert.Error(t, err)
}

func TestSelectBuilderPlaceholders(t *testing.T) {
	b := Select("test").PreWhere("x = ? AND y = ?").Where("x = ? AND y = ?")

	sql, _, _ := b.PlaceholderFormat(Question).ToSql()
	assert.Equal(t, "SELECT test PREWHERE x = ? AND y = ? WHERE x = ? AND y = ?", sql)

	sql, _, _ = b.PlaceholderFormat(Dollar).ToSql()
	assert.Equal(t, "SELECT test PREWHERE x = $1 AND y = $2 WHERE x = $3 AND y = $4", sql)

	sql, _, _ = b.PlaceholderFormat(Colon).ToSql()
	assert.Equal(t, "SELECT test PREWHERE x = :1 AND y = :2 WHERE x = :3 AND y = :4", sql)
}

func TestSelectBuilderRunners(t *testing.T) {
	db := &DBStub{}
	b := Select("test").RunWith(db)

	expectedSql := "SELECT test"

	b.Exec()
	assert.Equal(t, expectedSql, db.LastExecSql)

	b.Query()
	assert.Equal(t, expectedSql, db.LastQuerySql)

	b.QueryRow()
	assert.Equal(t, expectedSql, db.LastQueryRowSql)

	err := b.Scan()
	assert.NoError(t, err)
}

func TestSelectBuilderNoRunner(t *testing.T) {
	b := Select("test")

	_, err := b.Exec()
	assert.Equal(t, RunnerNotSet, err)

	_, err = b.Query()
	assert.Equal(t, RunnerNotSet, err)

	err = b.Scan()
	assert.Equal(t, RunnerNotSet, err)
}

func TestSelectBuilderSimpleJoin(t *testing.T) {

	expectedSql := "SELECT * FROM bar JOIN baz ON bar.foo = baz.foo"
	expectedArgs := []interface{}(nil)

	b := Select("*").From("bar").Join("baz ON bar.foo = baz.foo")

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	assert.Equal(t, expectedSql, sql)
	assert.Equal(t, args, expectedArgs)
}

func TestSelectBuilderParamJoin(t *testing.T) {

	expectedSql := "SELECT * FROM bar JOIN baz ON bar.foo = baz.foo AND baz.foo = ?"
	expectedArgs := []interface{}{42}

	b := Select("*").From("bar").Join("baz ON bar.foo = baz.foo AND baz.foo = ?", 42)

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	assert.Equal(t, expectedSql, sql)
	assert.Equal(t, args, expectedArgs)
}

func TestSelectBuilderNestedSelectJoin(t *testing.T) {

	expectedSql := "SELECT * FROM bar JOIN ( SELECT * FROM baz PREWHERE foo = ? WHERE foo = ? ) r ON bar.foo = r.foo"
	expectedArgs := []interface{}{42, 42}

	nestedSelect := Select("*").From("baz").PreWhere("foo = ?", 42).Where("foo = ?", 42)

	b := Select("*").From("bar").JoinClause(nestedSelect.Prefix("JOIN (").Suffix(") r ON bar.foo = r.foo"))

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	assert.Equal(t, expectedSql, sql)
	assert.Equal(t, args, expectedArgs)
}

func TestSelectWithOptions(t *testing.T) {
	sql, _, err := Select("*").From("foo").Distinct().Options("SQL_NO_CACHE").ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT DISTINCT SQL_NO_CACHE * FROM foo", sql)
}

func TestSelectWithRemoveLimit(t *testing.T) {
	sql, _, err := Select("*").From("foo").Limit(10).RemoveLimit().ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM foo", sql)
}

func TestSelectWithRemoveOffset(t *testing.T) {
	sql, _, err := Select("*").From("foo").Offset(10).RemoveOffset().ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM foo", sql)
}

func TestSelectBuilderNestedSelectDollar(t *testing.T) {
	nestedBuilder := StatementBuilder.PlaceholderFormat(Dollar).Select("*").Prefix("NOT EXISTS (").
		From("bar").PreWhere("y = ?", 42).Where("y = ?", 42).Suffix(")")
	outerSql, _, err := StatementBuilder.PlaceholderFormat(Dollar).Select("*").
		From("foo").PreWhere("x = ?").Where(nestedBuilder).ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM foo PREWHERE x = $1 WHERE NOT EXISTS ( SELECT * FROM bar PREWHERE y = $2 WHERE y = $3 )", outerSql)
}

func TestMustSql(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("TestUserFail should have panicked!")
			}
		}()
		// This function should cause a panic
		Select().From("foo").MustSql()
	}()
}

func TestSelectWithoutPreWhereOrWhereClause(t *testing.T) {
	sql, _, err := Select("*").From("users").ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM users", sql)
}

func TestSelectWithNilPreWhereNilWhereClause(t *testing.T) {
	sql, _, err := Select("*").From("users").PreWhere(nil).Where(nil).ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM users", sql)
}

func TestSelectWithEmptyStringPreWhereEmptyStringWhereClause(t *testing.T) {
	sql, _, err := Select("*").From("users").PreWhere("").Where("").ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM users", sql)
}
